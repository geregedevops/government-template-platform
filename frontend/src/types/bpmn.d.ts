// Ambient типийн зарлал — bpmn-js ба @bpmn-io/form-js нь өөрсдийн TypeScript
// тодорхойлолтыг түгээдэггүй тул бид ашигладаг хэсгийнх нь минимал интерфэйсийг
// энд зарлана (төслийн strict TS-д тохирно).

declare module 'bpmn-js/lib/Modeler' {
  export interface ModelerOptions {
    container?: HTMLElement | string | null;
    keyboard?: { bindTo?: Document | HTMLElement };
    additionalModules?: unknown[];
  }
  export interface ImportXMLResult {
    warnings: unknown[];
  }
  export interface SaveXMLResult {
    xml: string;
  }
  export default class BpmnModeler {
    constructor(options?: ModelerOptions);
    importXML(xml: string): Promise<ImportXMLResult>;
    saveXML(options?: { format?: boolean }): Promise<SaveXMLResult>;
    get<T = unknown>(name: string): T;
    destroy(): void;
  }
}

declare module 'bpmn-js/lib/NavigatedViewer' {
  export interface ViewerOptions {
    container?: HTMLElement | string | null;
  }
  export default class NavigatedViewer {
    constructor(options?: ViewerOptions);
    importXML(xml: string): Promise<{ warnings: unknown[] }>;
    get<T = unknown>(name: string): T;
    destroy(): void;
  }
}

declare module '@bpmn-io/form-js' {
  export interface FormOptions {
    container?: HTMLElement | string | null;
  }
  export interface FormSubmitResult {
    data: Record<string, unknown>;
    errors: Record<string, unknown>;
  }
  export class Form {
    constructor(options?: FormOptions);
    importSchema(schema: unknown, data?: Record<string, unknown>): Promise<unknown>;
    submit(): FormSubmitResult;
    on(event: string, callback: (event: FormSubmitResult) => void): void;
    destroy(): void;
  }
  export class FormEditor {
    constructor(options?: FormOptions);
    importSchema(schema: unknown): Promise<unknown>;
    saveSchema(): Record<string, unknown>;
    on(event: string, callback: (event: unknown) => void): void;
    destroy(): void;
  }
}

declare module '@bpmn-io/form-js-viewer' {
  export interface FormSubmitResult {
    data: Record<string, unknown>;
    errors: Record<string, unknown>;
  }
  export interface CreateFormOptions {
    schema: unknown;
    data?: Record<string, unknown>;
    container?: HTMLElement | string | null;
  }
  export interface FormViewerInstance {
    submit(): FormSubmitResult;
    on(event: string, callback: (event: FormSubmitResult) => void): void;
    destroy(): void;
  }
  export function createForm(options: CreateFormOptions): Promise<FormViewerInstance>;
}
