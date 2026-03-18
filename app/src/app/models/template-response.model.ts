import { TemplateMetadata } from './template.model';

export interface TemplatesResponse {
    templates: Record<string, TemplateMetadata>;
}