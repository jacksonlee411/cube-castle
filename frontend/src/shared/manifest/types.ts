export interface ManifestFieldMeta {
  name: string
  label: string
  description?: string
  placeholder?: string
  required?: boolean
}

export type ManifestFieldMap<T extends string> = Record<T, ManifestFieldMeta>
