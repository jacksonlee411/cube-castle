-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE public.standard_objects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_type text NOT NULL,
    code text NOT NULL,
    tenant_code text NOT NULL,
    display_name text NOT NULL,
    status text NOT NULL DEFAULT 'DRAFT',
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    schema_version text NOT NULL,
    data_classification text NOT NULL DEFAULT 'internal',
    retention_policy text NOT NULL DEFAULT 'standard',
    created_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_code, object_type, code)
);

CREATE TABLE public.standard_object_versions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    version_code text NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    validity_range tstzrange NOT NULL,
    transaction_range tstzrange NOT NULL,
    is_current boolean NOT NULL DEFAULT false,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    audit jsonb NOT NULL DEFAULT '{}'::jsonb,
    checksum text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT standard_object_versions_validity CHECK (
        lower(validity_range) <= upper(validity_range) OR upper_inf(validity_range)
    ),
    CONSTRAINT standard_object_versions_transaction CHECK (
        lower(transaction_range) <= upper(transaction_range) OR upper_inf(transaction_range)
    )
);

CREATE UNIQUE INDEX idx_standard_object_versions_version_code
    ON public.standard_object_versions(object_id, version_code);
CREATE UNIQUE INDEX idx_standard_object_versions_effective
    ON public.standard_object_versions(object_id, effective_date);
CREATE INDEX idx_standard_object_versions_current
    ON public.standard_object_versions(object_id)
    WHERE is_current;
CREATE INDEX idx_standard_object_versions_validity
    ON public.standard_object_versions USING GIST(object_id, validity_range);
CREATE INDEX idx_standard_object_versions_transaction
    ON public.standard_object_versions USING GIST(object_id, transaction_range);
ALTER TABLE public.standard_object_versions
    ADD CONSTRAINT standard_object_versions_valid_excl EXCLUDE USING gist (
        object_id WITH =,
        validity_range WITH &&
    );
ALTER TABLE public.standard_object_versions
    ADD CONSTRAINT standard_object_versions_tx_excl EXCLUDE USING gist (
        object_id WITH =,
        transaction_range WITH &&
    );

CREATE TABLE public.standard_object_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    link_type text NOT NULL,
    source_object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    target_object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    tenant_code text NOT NULL,
    validity_range tstzrange NOT NULL,
    transaction_range tstzrange NOT NULL,
    attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_by uuid,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT standard_object_links_validity CHECK (
        lower(validity_range) <= upper(validity_range) OR upper_inf(validity_range)
    ),
    CONSTRAINT standard_object_links_transaction CHECK (
        lower(transaction_range) <= upper(transaction_range) OR upper_inf(transaction_range)
    ),
    UNIQUE (link_type, source_object_id, target_object_id)
);
CREATE INDEX idx_standard_object_links_validity
    ON public.standard_object_links USING GIST(source_object_id, validity_range);
CREATE INDEX idx_standard_object_links_transaction
    ON public.standard_object_links USING GIST(source_object_id, transaction_range);
ALTER TABLE public.standard_object_links
    ADD CONSTRAINT standard_object_links_validity_excl EXCLUDE USING gist (
        source_object_id WITH =,
        validity_range WITH &&
    );
ALTER TABLE public.standard_object_links
    ADD CONSTRAINT standard_object_links_transaction_excl EXCLUDE USING gist (
        source_object_id WITH =,
        transaction_range WITH &&
    );

CREATE TABLE public.standard_object_hierarchy_snapshots (
    snapshot_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_code text NOT NULL,
    object_type text NOT NULL,
    as_of_valid timestamptz NOT NULL,
    as_of_transaction timestamptz NOT NULL,
    ancestor_object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    descendant_object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    depth integer NOT NULL,
    code_path text NOT NULL,
    name_path text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    refreshed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_standard_object_snapshots_tenant
    ON public.standard_object_hierarchy_snapshots(tenant_code, object_type, as_of_valid);

CREATE TABLE public.standard_object_schemas (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_type text NOT NULL,
    schema_version text NOT NULL,
    schema_hash text NOT NULL,
    definition jsonb NOT NULL DEFAULT '{}'::jsonb,
    dec_bindings jsonb NOT NULL DEFAULT '[]'::jsonb,
    ocl_guards jsonb NOT NULL DEFAULT '[]'::jsonb,
    glossary_url text,
    time_constraint text NOT NULL DEFAULT 'TC1',
    transaction_policy text NOT NULL DEFAULT 'APPEND_ONLY',
    published_at timestamptz NOT NULL DEFAULT now(),
    rollback_version text,
    maintainer text,
    CONSTRAINT standard_object_schemas_unique UNIQUE (object_type, schema_version),
    CONSTRAINT standard_object_schemas_time_constraint CHECK (
        time_constraint = ANY (ARRAY['TC1','TC2','TC3'])
    ),
    CONSTRAINT standard_object_schemas_transaction_policy CHECK (
        transaction_policy = ANY (ARRAY['APPEND_ONLY','CORRECTION_ALLOWED'])
    )
);

CREATE TABLE public.standard_object_translations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    tenant_code text NOT NULL,
    locale text NOT NULL,
    display_name text,
    description text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (object_id, locale)
);

CREATE TABLE public.standard_object_attachments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    tenant_code text NOT NULL,
    attachment_type text NOT NULL,
    uri text NOT NULL,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by uuid,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_standard_object_attachments_object
    ON public.standard_object_attachments(object_id);

CREATE TABLE public.standard_object_metadata (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid NOT NULL REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    tenant_code text NOT NULL,
    meta_key text NOT NULL,
    meta_value jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (object_id, meta_key)
);

CREATE TABLE public.standard_object_metrics (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id uuid REFERENCES public.standard_objects(id) ON DELETE CASCADE,
    metric_type text NOT NULL,
    metric_value numeric NOT NULL,
    labels jsonb NOT NULL DEFAULT '{}'::jsonb,
    tenant_code text NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_standard_object_metrics_object
    ON public.standard_object_metrics(object_id, metric_type);
CREATE INDEX idx_standard_object_metrics_recorded
    ON public.standard_object_metrics(recorded_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.standard_object_metrics;
DROP TABLE IF EXISTS public.standard_object_metadata;
DROP TABLE IF EXISTS public.standard_object_attachments;
DROP TABLE IF EXISTS public.standard_object_translations;
DROP TABLE IF EXISTS public.standard_object_schemas;
DROP TABLE IF EXISTS public.standard_object_hierarchy_snapshots;
DROP TABLE IF EXISTS public.standard_object_links;
DROP TABLE IF EXISTS public.standard_object_versions;
DROP TABLE IF EXISTS public.standard_objects;
-- +goose StatementEnd
