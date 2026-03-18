export interface Template {
    filename: string;
    info: Info;
    template_dir: string;
    steps: Step[];
    hooks: HookConfig;
    global_vars?: Record<string, any>;
}

export interface Info {
    name: string;
    author: string;
    description?: string;
    category?: string;
    system: boolean;
    tags?: string[];
}

export interface Step {
    id: string;
    title: string;
    path: string;
    method: string;
    template_file: string;
    success_message?: string;
    next?: string;
    redirect_url?: string;
    capture: CaptureConfig;
    simulate_delay_ms?: number;
    vars?: Record<string, any>;
}

export interface CaptureConfig {
    enabled: boolean;
    fields?: CaptureField[];
}

export interface CaptureField {
    name: string;
    required: boolean;
    validate_regex?: string;
    error_message?: string;
}

export interface HookConfig {
    onLoad?: string[];
}

export interface HtmlFile {
    filename: string;
    path: string;
    size: number;
    mod_time: string;
}

export interface TemplateMetadata {
    content: string;
    filename: string;
    name: string;
    author: string;
    description?: string;
    category?: string;
    tags?: string[];
    info: Info;
    template_dir: string;
    size: number;
    mod_time: string;
    is_dir: boolean;
    mode: string;
    html_files: HtmlFile[];
}

export interface TemplateCreateRequest {
    filename: string;
    content: string;
}

export interface TemplateUpdateRequest {
    filename: string;
    content: string;
}

export interface TemplateDeleteRequest {
    filename: string;
}


export interface TemplateHtmlFile {
    filename: string;
    path: string;
    size: number;
    mod_time: string;
    content: string;
}

export interface TemplateHtmlFileUpdateRequest {
    t_filename: string;
    filename: string;
    content: string;
}

export interface TemplateHtmlFileUploadRequest {
    t_filename: string;
    filename: string;
    content: string;
}

export interface TemplateHtmlFileDeleteRequest {
    t_filename: string;
    filename: string;
}

export interface TemplateStaticFileRequest {
    filename: string;
    t_filename: string;
    content: string;
}

export interface TemplateStaticFile {
    filename: string;
    path: string;
    size: number;
    mod_time: string;
    content: string;
}