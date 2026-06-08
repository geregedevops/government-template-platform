// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "geregetemplateai/docs" // swagger тодорхойлолт, swaggo-оор init үед бүртгэгддэг
	aiusecase "geregetemplateai/internal/business/usecases/ai"
	"geregetemplateai/internal/business/usecases/auth"
	bpmusecase "geregetemplateai/internal/business/usecases/bpm"
	fedusecase "geregetemplateai/internal/business/usecases/federation"
	orgusecase "geregetemplateai/internal/business/usecases/organization"
	rbacusecase "geregetemplateai/internal/business/usecases/rbac"
	"geregetemplateai/internal/business/usecases/users"
	voiceusecase "geregetemplateai/internal/business/usecases/voice"
	"geregetemplateai/internal/config"
	"geregetemplateai/internal/constants"
	"geregetemplateai/internal/datasources/caches"
	"geregetemplateai/internal/datasources/drivers"
	aipostgres "geregetemplateai/internal/datasources/repositories/postgres/ai"
	bpmpostgres "geregetemplateai/internal/datasources/repositories/postgres/bpm"
	fedpostgres "geregetemplateai/internal/datasources/repositories/postgres/federation"
	orgpostgres "geregetemplateai/internal/datasources/repositories/postgres/organization"
	rbacpostgres "geregetemplateai/internal/datasources/repositories/postgres/rbac"
	userspostgres "geregetemplateai/internal/datasources/repositories/postgres/users"
	voicepostgres "geregetemplateai/internal/datasources/repositories/postgres/voice"
	V1Handler "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/internal/http/middlewares"
	"geregetemplateai/internal/http/routes"
	"geregetemplateai/pkg/aiclient"
	"geregetemplateai/pkg/bpmconnector"
	"geregetemplateai/pkg/fedsign"
	"geregetemplateai/pkg/geminiclient"
	"geregetemplateai/pkg/jwt"
	"geregetemplateai/pkg/logger"
	"geregetemplateai/pkg/mailer"
	"geregetemplateai/pkg/observability"
	"geregetemplateai/pkg/verify"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

const serviceName = "gerege-template"

// bpmClaudeGenerator нь bpmusecase.Generator-ийг Claude (aiclient.Streamer)
// дээр хэрэгжүүлнэ — stream-ийг бүтэн текст болгож хураана (generate нь
// streaming биш, нэг JSON хариу).
type bpmClaudeGenerator struct{ streamer aiclient.Streamer }

func (g bpmClaudeGenerator) Generate(ctx context.Context, system, userMessage string) (string, error) {
	full, _, err := g.streamer.StreamMessage(ctx, system,
		[]aiclient.Message{{Role: "user", Content: userMessage}},
		func(string) error { return nil },
	)
	return full, err
}

type App struct {
	fiber            *fiber.App
	db               *gorm.DB
	redisCache       caches.RedisCache
	asyncMailer      *mailer.AsyncOTPMailer
	tracerShutdown   observability.Shutdown
	authRateLimiter  *middlewares.RateLimiter
	aiRateLimiter    *middlewares.RateLimiter
	voiceRateLimiter *middlewares.RateLimiter
	bpmRateLimiter   *middlewares.RateLimiter
	fedWorkerStop    chan struct{}
}

func NewApp() (*App, error) {
	// Tracer-ийг эхэлд тохируулна — ингэснээр дараагийн тохиргооноос (DB холболт,
	// migration шалгалт гэх мэт) ялгарах span-ууд зөв provider руу очно.
	shutdownTracer, err := observability.SetupTracing(context.Background(), observability.TracingConfig{
		ServiceName: serviceName,
		Environment: config.AppConfig.Environment,
		Exporter:    config.AppConfig.OTelExporter,
		SampleRatio: config.AppConfig.OTelSampleRatio,
	})
	if err != nil {
		return nil, fmt.Errorf("setup tracing: %w", err)
	}

	// өгөгдлийн сангуудыг тохируулах
	conn, err := drivers.SetupGORMPostgres()
	if err != nil {
		return nil, err
	}
	// DB pool-ийн бодит статистикийг /metrics-ээр гаргана — provider нь GORM
	// handle-аас авсан түүхий *sql.DB-г буцаана.
	observability.RegisterDBStatsProvider(func() *sql.DB {
		sqlDB, dbErr := conn.DB()
		if dbErr != nil {
			return nil
		}
		return sqlDB
	})

	// jwt сервис
	jwtService := jwt.NewJWTServiceWithRefresh(
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTIssuer,
		config.AppConfig.JWTExpired,
		config.AppConfig.JWTRefreshExpired,
	)
	// ES256 (asymmetric) горим — EC хувийн түлхүүр тохируулсан бол. Федерацийн
	// бусад node токеныг /.well-known/jwks.json-оор шалгана. Тохируулаагүй бол
	// HS256 (анхдагч) хэвээр — одоо ажиллаж буй node-д өөрчлөлтгүй.
	if config.AppConfig.JWTECPrivateKey != "" {
		es, ecErr := jwt.EnableES256(jwtService, config.AppConfig.JWTECPrivateKey, config.AppConfig.JWTECKid)
		if ecErr != nil {
			return nil, fmt.Errorf("enable ES256: %w", ecErr)
		}
		jwtService = es
		logger.Info("jwt: ES256 (asymmetric) mode enabled; JWKS published", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
	}

	// кэш
	redisCache := caches.NewRedisCache(config.AppConfig.REDISHost, 0, config.AppConfig.REDISPassword, time.Duration(config.AppConfig.REDISExpired))
	ristrettoCache, err := caches.NewRistrettoCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	// mailer — синхрон SMTP илгээгчийг асинхрон дараалалд (queue) боож,
	// OTP илгээх хоцролтыг HTTP хүсэлтийн замаас гаргана.
	syncMailer := mailer.NewOTPMailer(config.AppConfig.OTPEmail, config.AppConfig.OTPPassword)
	asyncMailer := mailer.NewAsyncOTPMailer(
		syncMailer,
		config.AppConfig.MailerWorkers,
		config.AppConfig.MailerQueueSize,
		config.AppConfig.MailerRetries,
		time.Second,
	)

	// router + глобал middleware-г тохируулах
	app := setupRouter()

	// auth middleware — хүчинтэй токентой хэрэглэгч endpoint-д хандах боломжтой.
	// redisCache-г дамжуулсан тул middleware нь ChangePassword / ResetPassword-оор
	// бичигдсэн нууц үг солих хугацааны хязгаарыг (cutoff) баримтлаж чадна.
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, redisCache, false)

	// Дэд бүтцийн (infrastructure) endpoint-ууд (/api бүлгээс гадуур)
	healthHandler := V1Handler.NewHealthHandler(conn, redisCache.Client())
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Ready)
	// JWKS — ES256 горимд нийтийн түлхүүрийг нийтэлнэ (федерацийн node-ууд
	// Gerege-ийн токеныг үүгээр баталгаажуулна). HS256 горимд хоосон keys.
	app.Get("/.well-known/jwks.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		if jwks, ok := jwtService.JWKSet(); ok {
			return c.Send(jwks)
		}
		return c.SendString(`{"keys":[]}`)
	})
	// /metrics, /swagger/doc.json нь production-д systeметрик ба API
	// гадаргууг ил гаргадаг тул өндөр эрсдэлтэй. Гате-чин (gate) middleware:
	//   - dev орчинд token-гүй бол ил байна (`make serve` UX-ийг хадгална);
	//   - prod орчинд OBSERVABILITY_TOKEN хоосон бол 404 буцаана
	//     (endpoint огт байхгүйгээр харагдана);
	//   - prod орчинд OBSERVABILITY_TOKEN тохируулсан бол "Authorization:
	//     Bearer <token>" нийлэхэд л зөвшөөрнө.
	obsGate := middlewares.ObservabilityGate(
		config.AppConfig.Environment == constants.EnvironmentProduction,
		config.AppConfig.ObservabilityToken,
	)
	// /metrics — Prometheus exposition. promhttp нь net/http handler бөгөөд
	// adaptor middleware-ээр дамжуулан Fiber руу холбогддог.
	app.Get("/metrics", obsGate, adaptor.HTTPHandler(promhttp.Handler()))
	// OpenAPI тодорхойлолт — `make swag` нь godoc annotation-уудаас docs/
	// багцыг үүсгэдэг. gofiber/swagger нь Fiber v2-д зориулагдсан тул
	// (handler нь *fiber.Ctx авдаг) Fiber v3-д runtime panic үүсгэдэг —
	// иймд spec-ийг Fiber v3 native-аар JSON хэлбэрээр үйлчилнэ. Уг JSON-ыг
	// Swagger UI / Postman / VS Code-д шууд ачаалж болно. (Суулгасан
	// интерактив UI хэрэгтэй бол Fiber v3-тэй нийцэх swagger handler нэмнэ.)
	app.Get("/swagger/doc.json", obsGate, func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(docs.SwaggerInfo.ReadDoc())
	})

	// Хязгаарлагдсан контекстуудыг (bounded context) угсарна. Users нь
	// identity CRUD-г эзэмшиж, Auth нь credential / session урсгалыг эзэмшинэ;
	// Auth нь хэрэглэгчийн бичлэгийг унших/бичихдээ Users-ээс хамаардаг.
	userRepo := userspostgres.NewUserRepository(conn)
	usersUC := users.NewUsecase(userRepo, ristrettoCache, users.Config{
		BcryptCost: config.AppConfig.BcryptCost,
	})
	// GeregeCloud Verify клиент — OTP илгээх/шалгах ажлыг алсын үйлчилгээнд
	// шилжүүлнэ. VerifyAPIKey хоосон бол клиент бүтэх боловч дуудлага бүр
	// "missing api key" алдаа буцаах тул operator-д чимээгүй буруу тохиргоо
	// үлдэхгүй.
	verifyClient := verify.NewClient(
		config.AppConfig.VerifyAPIBase,
		config.AppConfig.VerifyAPIKey,
		config.AppConfig.VerifyChannel,
	)
	authUC := auth.NewUsecase(usersUC, jwtService, asyncMailer, verifyClient, redisCache, auth.Config{
		OTPMaxAttempts:         config.AppConfig.OTPMaxAttempts,
		OTPTTL:                 time.Duration(config.AppConfig.REDISExpired) * time.Minute,
		PasswordResetTTL:       30 * time.Minute,
		BcryptCost:             config.AppConfig.BcryptCost,
		LoginMaxAttempts:       10,
		GlobalLoginMaxAttempts: 100, // email-ийн бүх IP-ийн нийт босго (тархсан brute-force)
		LoginLockoutTTL:        15 * time.Minute,
		ForgotMaxAttempts:      3,
		ForgotLockoutTTL:       15 * time.Minute,
	})

	// Эрх (role) солигдоход тухайн хэрэглэгчийн токеныг хүчингүй болгож шинэ
	// эрхийг нэн даруй хүчинтэй болгохын тулд authUC-ийг users-д TokenRevoker
	// болгон холбоно (users нь auth-аас өмнө үүсдэг тул setter-ээр).
	if revoker, ok := authUC.(users.TokenRevoker); ok {
		if setter, ok2 := usersUC.(interface{ SetTokenRevoker(users.TokenRevoker) }); ok2 {
			setter.SetTokenRevoker(revoker)
		}
	}

	// Нэргүй /auth гадаргуун дээр IP тус бүрт минутанд 5 хүсэлт зөвшөөрнө.
	// App нь түүнийг эзэмшдэг тул фоны цэвэрлэгээний goroutine-ийг
	// graceful shutdown (эвсэг унтраалт) үед зогсоож болно.
	authRateLimiter := middlewares.NewRateLimiter(rate.Limit(5.0/60.0), 5)

	// AI туслах (Claude). Түлхүүр тохируулаагүй үед endpoint-ууд 503
	// буцаана — route-уудыг нөхцөлтэйгөөр mount хийхгүй (404 vs 503-ийн
	// ялгаа нь operator-т тохиргооны алдааг шууд харуулна).
	claudeClient := aiclient.NewClient(
		config.AppConfig.AnthropicAPIKey,
		config.AppConfig.AnthropicModel,
		config.AppConfig.AIMaxTokens,
		time.Duration(config.AppConfig.AIRequestTimeoutSecs)*time.Second,
	)
	aiRepo := aipostgres.NewAIRepository(conn)
	aiUC := aiusecase.NewUsecase(aiRepo, claudeClient, redisCache, aiusecase.Config{
		Enabled:           claudeClient.Configured(),
		Model:             claudeClient.Model(),
		DailyRequestLimit: config.AppConfig.AIDailyRequestLimit,
		HistoryLimit:      config.AppConfig.AIHistoryLimit,
	})
	// AI гадаргуу нь баталгаажсан хэрэглэгчдэд л нээлттэй тул auth-аас
	// зөөлөн (IP тус бүрт минутанд 20) хязгаартай — streaming хүсэлтүүд
	// удаан үргэлжилдэг тул хэт чанга тоолуур UX-ийг эвдэнэ.
	aiRateLimiter := middlewares.NewRateLimiter(rate.Limit(20.0/60.0), 10)

	// Дуу хоолойн орчуулга (Gemini, MN↔EN). AI-тэй ижил: түлхүүр
	// тохируулаагүй үед endpoint-ууд 503 буцаана (route нөхцөлтэйгөөр
	// mount хийхгүй). Нэг хүсэлт Gemini руу 2 дуудлага (STT/орчуулга + TTS)
	// хийдэг тул AI-аас чанга хязгаар — IP тус бүрт минутанд 10.
	geminiClient := geminiclient.NewClient(
		config.AppConfig.GeminiAPIKey,
		config.AppConfig.GeminiModel,
		config.AppConfig.GeminiTTSModel,
		config.AppConfig.GeminiVoice,
		time.Duration(config.AppConfig.VoiceRequestTimeoutSecs)*time.Second,
	)
	voiceRepo := voicepostgres.NewVoiceRepository(conn)
	voiceUC := voiceusecase.NewUsecase(voiceRepo, geminiClient, redisCache, voiceusecase.Config{
		Enabled:           geminiClient.Configured(),
		Model:             geminiClient.Model(),
		DailyRequestLimit: config.AppConfig.VoiceDailyRequestLimit,
		MaxAudioBytes:     config.AppConfig.VoiceMaxAudioKB * 1024,
	})
	voiceRateLimiter := middlewares.NewRateLimiter(rate.Limit(10.0/60.0), 5)

	// BPM (Business Process Management). AI/voice-аас ялгаатай нь гадаад
	// провайдер шаардахгүй — процессын CRUD + хөнгөн гүйлт нь зөвхөн DB-д
	// тулгуурладаг. Зөв тохиргоо нь auth-тай ижил нягт (IP тус бүрт минутанд
	// 30 — modeler нь хадгалах/жагсаах олон хүсэлт хийж болзошгүй).
	bpmRepo := bpmpostgres.NewBPMRepository(conn)
	// serviceTask нь хэрэглэгчийн тохируулсан URL руу HTTP дуудлага хийдэг тул
	// connector нь SSRF-хамгаалалттай (хувийн/loopback IP-г хориглоно).
	bpmConnector := bpmconnector.New(10*time.Second, 1<<20)
	// AI-аар процесс үүсгэх — Claude-ийн тусдаа клиент (илүү том token budget,
	// spec JSON таслагдахгүй). Түлхүүр тохируулаагүй бол generator nil →
	// /bpm/generate нь 503 буцаана.
	var bpmGenerator bpmusecase.Generator
	if claudeClient.Configured() {
		bpmGenClient := aiclient.NewClient(
			config.AppConfig.AnthropicAPIKey,
			config.AppConfig.AnthropicModel,
			// Олон node + маягттай процессын JSON spec нь том байж болзошгүй тул
			// 4096 хүрэлцэхгүй (тасарч JSON эвдэрнэ). Уужим хязгаар тавина.
			16384,
			time.Duration(config.AppConfig.AIRequestTimeoutSecs)*time.Second,
		)
		bpmGenerator = bpmClaudeGenerator{streamer: bpmGenClient}
	}
	bpmUC := bpmusecase.NewUsecase(bpmRepo, bpmConnector, bpmGenerator, redisCache, bpmusecase.Config{
		AIEnabled: claudeClient.Configured(),
		// AI-аар процесс үүсгэх нь токен зарцуулдаг тул AI чатын адил өдрийн
		// лимит хэрэглэнэ (нэг env-ийг хуваалцана).
		GenerateDailyLimit: config.AppConfig.AIDailyRequestLimit,
	})
	bpmRateLimiter := middlewares.NewRateLimiter(rate.Limit(30.0/60.0), 15)

	// RBAC: динамик эрх + permission enforcement. rbacUC нь middleware-ийн
	// PermissionResolver-ийг хангадаг тул feature route бүрд permission шалгалт
	// тавина.
	rbacRepo := rbacpostgres.NewRBACRepository(conn)
	rbacUC := rbacusecase.NewUsecase(rbacRepo)

	// Байгууллагын мод (ROADMAP Үе 0 / P0) — федератив hierarchy-ийн суурь.
	orgRepo := orgpostgres.NewOrganizationRepository(conn)
	orgUC := orgusecase.NewUsecase(orgRepo)

	// Федераци (ROADMAP Үе 1) — node-ийн e-seal (FED EC түлхүүр) тохируулсан бол
	// гарын үсэг идэвхжинэ; эс бөгөөс signer=nil (федераци идэвхгүй, бусад
	// функцэд нөлөөгүй).
	var fedSigner *fedsign.Signer
	if config.AppConfig.FedECPrivateKey != "" {
		s, fsErr := fedsign.NewSigner(config.AppConfig.FedECPrivateKey, config.AppConfig.FedECKid, config.AppConfig.FedNodeID)
		if fsErr != nil {
			return nil, fmt.Errorf("federation signer: %w", fsErr)
		}
		fedSigner = s
		logger.Info("federation: node identity enabled; fed-jwks published", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryServer})
	}
	fedRepo := fedpostgres.NewFederationRepository(conn)
	fedUC := fedusecase.NewUsecase(fedRepo, fedSigner)
	// fed-JWKS — node-ийн гарын үсгийн нийтийн түлхүүр (peer-ууд мессеж шалгана).
	app.Get("/.well-known/fed-jwks.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		if jwks, ok := fedUC.JWKS(); ok {
			return c.Send(jwks)
		}
		return c.SendString(`{"keys":[]}`)
	})
	// delegatedTask-ийн bpm↔federation мөчлөгийг интерфейсээр таслаж холбоно
	// (аль алийг үүсгэсний дараа).
	bpmUC.SetFedSender(fedUC)
	fedUC.SetDelegationHandler(bpmUC)

	// API Route-ууд
	api := app.Group("/api")
	api.Get("/", routes.RootHandler)
	routes.NewAuthRoute(api, authUC, authMiddleware, authRateLimiter).Routes()
	routes.NewUsersRoute(api, usersUC, authMiddleware, rbacUC).Routes()
	routes.NewRBACRoute(api, rbacUC, authMiddleware).Routes()
	routes.NewOrganizationRoute(api, orgUC, authMiddleware, rbacUC).Routes()
	routes.NewAIRoute(api, aiUC, authMiddleware, aiRateLimiter,
		time.Duration(config.AppConfig.AIRequestTimeoutSecs)*time.Second, rbacUC).Routes()
	routes.NewVoiceRoute(api, voiceUC, authMiddleware, voiceRateLimiter, rbacUC).Routes()
	routes.NewBPMRoute(api, bpmUC, authMiddleware, bpmRateLimiter, rbacUC).Routes()
	routes.NewFederationRoute(api, fedUC, authMiddleware, rbacUC).Routes()

	// Федерацийн outbox worker — гарах гарын үсэгтэй мессежийг найдвартай
	// (backoff retry-той) хүргэнэ. Зөвхөн node identity тохируулсан үед.
	fedWorkerStop := make(chan struct{})
	if fedUC.Configured() {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-fedWorkerStop:
					return
				case <-ticker.C:
					fedUC.ProcessOutbox(context.Background())
				}
			}
		}()
	}

	return &App{
		fiber:            app,
		db:               conn,
		redisCache:       redisCache,
		asyncMailer:      asyncMailer,
		tracerShutdown:   shutdownTracer,
		authRateLimiter:  authRateLimiter,
		aiRateLimiter:    aiRateLimiter,
		voiceRateLimiter: voiceRateLimiter,
		bpmRateLimiter:   bpmRateLimiter,
		fedWorkerStop:    fedWorkerStop,
	}, nil
}

func (a *App) Run() (err error) {
	srvLog := logger.WithFields(logger.Fields{constants.LoggerCategory: constants.LoggerCategoryServer})

	addr := fmt.Sprintf(":%d", config.AppConfig.Port)
	go func() {
		srvLog.Infof("success to listen and serve on %s", addr)
		if listenErr := a.fiber.Listen(addr); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			srvLog.Fatalf("Failed to listen and serve: %+v", listenErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	srvLog.Info("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Шинэ холболт хүлээж авахаа болиод, явагдаж буй хүсэлтүүдийг гүйцээнэ.
	if shutdownErr := a.fiber.ShutdownWithContext(ctx); shutdownErr != nil {
		return fmt.Errorf("error when shutdown server: %v", shutdownErr)
	}

	// Rate limiter-ийн цэвэрлэгээний goroutine-ийг зогсооно. Shutdown-ийн
	// дараа дуудахад аюулгүй — middleware руу шинэ хүсэлт ирэхгүй.
	if a.authRateLimiter != nil {
		a.authRateLimiter.Stop()
	}
	if a.aiRateLimiter != nil {
		a.aiRateLimiter.Stop()
	}
	if a.voiceRateLimiter != nil {
		a.voiceRateLimiter.Stop()
	}
	if a.bpmRateLimiter != nil {
		a.bpmRateLimiter.Stop()
	}
	if a.fedWorkerStop != nil {
		close(a.fedWorkerStop)
	}

	// хамаарлуудыг устгахаас өмнө асинхрон mailer-ийн дарааллыг гүйцээж,
	// явагдаж буй OTP имэйлүүд хүргэгдэх боломжтой болгоно.
	if a.asyncMailer != nil {
		if err := a.asyncMailer.Shutdown(ctx); err != nil {
			srvLog.Errorf("mailer shutdown incomplete: %v", err)
		}
	}

	// өгөгдлийн сангийн холболтыг хаах
	if sqlDB, dbErr := a.db.DB(); dbErr == nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			srvLog.Errorf("error closing database: %v", closeErr)
		}
	}

	// redis холболтыг хаах
	if err := a.redisCache.Close(); err != nil {
		srvLog.Errorf("error closing redis: %v", err)
	}

	// batch exporter-ийн буферлэж буй бүх span-уудыг flush хийнэ — HTTP сервер
	// хүсэлт хүлээж авахаа больсны дараа боловч процесс гарахаас өмнө ажиллах
	// ёстой, эс бөгөөс явагдаж буй trace-ийн төгсгөлийн хэсэг алдагдана.
	if a.tracerShutdown != nil {
		if err := a.tracerShutdown(ctx); err != nil {
			srvLog.Errorf("tracer shutdown incomplete: %v", err)
		}
	}

	srvLog.Info("server exiting")
	return
}

// setupRouter нь Fiber app-ыг бүтээж, глобал middleware стекийг суулгана.
// Дараалал чухал: эхэлд tracing — ингэснээр RequestIDMiddleware түүнийг
// logger context руу холбохоос өмнө span context (trace_id) тогтоогддог;
// стекийн доош ялгарах span-ууд (DB, Redis) автоматаар серверийн span-ийн
// дэд (child) болдог.
func setupRouter() *fiber.App {
	fiberCfg := fiber.Config{
		// Framework түвшний body-ийн дээд хязгаар — хамгаалалтын эхний шугам.
		// Route тус бүрийн илүү чанга хязгаарыг BodySizeLimitMiddleware-ээр тавина.
		BodyLimit: int(middlewares.DefaultBodyMaxBytes),
		// Төвлөрсөн алдааны handler: handler-ийн буцаасан аливаа алдаа (эсвэл
		// дээд талд сэргээгдсэн panic) энд цуглуулагдаж, нэгдсэн BaseResponse
		// дугтуй (envelope)-аар дүрслэгдэнэ.
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return V1Handler.RespondWithError(c, err)
		},
	}
	// Reverse proxy ард байх үед (nginx, ALB, Cloudflare г.м.) X-Forwarded-For-ийг
	// итгэлтэйгээр уншихын тулд TRUSTED_PROXIES-ийг тохируул. Тохиргоогүй үед
	// Fiber нь спуфинг хийсэн толгойг үл тоомсорлоод TCP peer-ийн IP-г буцаана —
	// энэ нь dev-д зөв (proxy байхгүй), харин production-д хууль ёсны клиентүүд
	// бүгд proxy-ийн ганц IP харагдах тул rate limit / audit / access log
	// эвдэрнэ. Operator оруулсан үед EnableIPValidation ороод header утга нь
	// үнэхээр IP байгаа эсэхийг шалгана.
	if proxies := config.AppConfig.TrustedProxiesList(); len(proxies) > 0 {
		fiberCfg.TrustProxy = true
		fiberCfg.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: proxies}
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
		fiberCfg.EnableIPValidation = true
	}
	app := fiber.New(fiberCfg)

	app.Use(middlewares.TracingMiddleware(serviceName))
	app.Use(middlewares.RequestIDMiddleware())
	// Accept-Language → mn/en. RequestID-ийн дараа, бусад бүхнээс өмнө —
	// ингэснээр аль ч давхаргын (rate limit 429 хүртэл) хариу мессеж
	// хэрэглэгчийн хэлээр орчуулагдах боломжтой.
	app.Use(middlewares.LocaleMiddleware())
	app.Use(middlewares.MetricsMiddleware())
	app.Use(middlewares.SecurityHeadersMiddleware())
	app.Use(middlewares.CORSMiddleware())
	app.Use(middlewares.BodySizeLimitMiddleware(middlewares.DefaultBodyMaxBytes))
	app.Use(middlewares.AccessLogMiddleware())
	// Хүсэлт бүрт deadline — гацсан handler/query холболтыг хэт удаан
	// эзлэхээс сэргийлнэ (secure_system_guide §5.3). tracing/request-id-ийн
	// дараа байрлуулсан тул context-ийн утгууд хадгалагдана.
	app.Use(middlewares.TimeoutMiddleware(middlewares.DefaultRequestTimeout))

	return app
}
