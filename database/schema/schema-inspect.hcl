table "audit_logs" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "event_type" {
    null = false
    type = character_varying(50)
  }
  column "resource_type" {
    null = false
    type = character_varying(50)
  }
  column "resource_id" {
    null = true
    type = character_varying(100)
  }
  column "actor_id" {
    null = true
    type = character_varying(100)
  }
  column "actor_type" {
    null = true
    type = character_varying(50)
  }
  column "action_name" {
    null = true
    type = character_varying(100)
  }
  column "request_id" {
    null = true
    type = character_varying(100)
  }
  column "operation_reason" {
    null = true
    type = text
  }
  column "timestamp" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "success" {
    null    = false
    type    = boolean
    default = true
  }
  column "error_code" {
    null = true
    type = character_varying(100)
  }
  column "error_message" {
    null = true
    type = text
  }
  column "request_data" {
    null    = false
    type    = jsonb
    default = "{}"
  }
  column "response_data" {
    null    = false
    type    = jsonb
    default = "{}"
  }
  column "modified_fields" {
    null    = false
    type    = jsonb
    default = "[]"
  }
  column "changes" {
    null    = false
    type    = jsonb
    default = "[]"
  }
  column "record_id" {
    null    = true
    type    = uuid
    comment = "组织单元时态版本的唯一标识，用于精确审计查询"
  }
  column "business_context" {
    null    = false
    type    = jsonb
    default = "{}"
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_audit_logs_record_id_time" {
    on {
      column = column.record_id
    }
    on {
      desc   = true
      column = column.timestamp
    }
  }
  index "idx_audit_logs_resource" {
    columns = [column.resource_type, column.resource_id]
  }
  index "idx_audit_logs_resource_timestamp" {
    on {
      column = column.resource_type
    }
    on {
      column = column.resource_id
    }
    on {
      desc   = true
      column = column.timestamp
    }
  }
  index "idx_audit_logs_timestamp" {
    columns = [column.timestamp]
  }
  check "audit_logs_event_type_check_v2" {
    expr = "((event_type)::text = ANY ((ARRAY['CREATE'::character varying, 'UPDATE'::character varying, 'DELETE'::character varying, 'SUSPEND'::character varying, 'REACTIVATE'::character varying, 'QUERY'::character varying, 'VALIDATION'::character varying, 'AUTHENTICATION'::character varying, 'ERROR'::character varying])::text[]))"
  }
}
table "goose_db_version" {
  schema = schema.public
  column "id" {
    null = false
    type = integer
    identity {
      generated = BY_DEFAULT
    }
  }
  column "version_id" {
    null = false
    type = bigint
  }
  column "is_applied" {
    null = false
    type = boolean
  }
  column "tstamp" {
    null    = false
    type    = timestamp
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
}
table "job_families" {
  schema = schema.public
  column "record_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "family_code" {
    null = false
    type = character_varying(20)
  }
  column "family_group_code" {
    null = false
    type = character_varying(20)
  }
  column "parent_record_id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  column "description" {
    null = true
    type = text
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "ACTIVE"
  }
  column "effective_date" {
    null = false
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.record_id]
  }
  foreign_key "fk_job_families_group" {
    columns     = [column.parent_record_id, column.tenant_id]
    ref_columns = [table.job_family_groups.column.record_id, table.job_family_groups.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "job_families_record_id_tenant_id_key" {
    unique  = true
    columns = [column.record_id, column.tenant_id]
  }
  index "job_families_tenant_id_family_code_effective_date_key" {
    unique  = true
    columns = [column.tenant_id, column.family_code, column.effective_date]
  }
  index "uk_job_families_current" {
    unique  = true
    columns = [column.tenant_id, column.family_code]
    where   = "(is_current = true)"
  }
  index "uk_job_families_record" {
    unique  = true
    columns = [column.record_id, column.tenant_id, column.family_code]
  }
}
table "job_family_groups" {
  schema = schema.public
  column "record_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "family_group_code" {
    null = false
    type = character_varying(20)
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  column "description" {
    null = true
    type = text
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "ACTIVE"
  }
  column "effective_date" {
    null = false
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.record_id]
  }
  index "job_family_groups_record_id_tenant_id_key" {
    unique  = true
    columns = [column.record_id, column.tenant_id]
  }
  index "job_family_groups_tenant_id_family_group_code_effective_dat_key" {
    unique  = true
    columns = [column.tenant_id, column.family_group_code, column.effective_date]
  }
  index "uk_job_family_groups_current" {
    unique  = true
    columns = [column.tenant_id, column.family_group_code]
    where   = "(is_current = true)"
  }
  index "uk_job_family_groups_record" {
    unique  = true
    columns = [column.record_id, column.tenant_id, column.family_group_code]
  }
}
table "job_levels" {
  schema = schema.public
  column "record_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "level_code" {
    null = false
    type = character_varying(20)
  }
  column "role_code" {
    null = false
    type = character_varying(20)
  }
  column "parent_record_id" {
    null = false
    type = uuid
  }
  column "level_rank" {
    null = false
    type = character_varying(20)
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  column "description" {
    null = true
    type = text
  }
  column "salary_band" {
    null = true
    type = jsonb
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "ACTIVE"
  }
  column "effective_date" {
    null = false
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.record_id]
  }
  foreign_key "fk_job_levels_role" {
    columns     = [column.parent_record_id, column.tenant_id]
    ref_columns = [table.job_roles.column.record_id, table.job_roles.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "job_levels_record_id_tenant_id_key" {
    unique  = true
    columns = [column.record_id, column.tenant_id]
  }
  index "job_levels_tenant_id_level_code_effective_date_key" {
    unique  = true
    columns = [column.tenant_id, column.level_code, column.effective_date]
  }
  index "uk_job_levels_current" {
    unique  = true
    columns = [column.tenant_id, column.level_code]
    where   = "(is_current = true)"
  }
  index "uk_job_levels_record" {
    unique  = true
    columns = [column.record_id, column.tenant_id, column.level_code]
  }
}
table "job_roles" {
  schema = schema.public
  column "record_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "role_code" {
    null = false
    type = character_varying(20)
  }
  column "family_code" {
    null = false
    type = character_varying(20)
  }
  column "parent_record_id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  column "description" {
    null = true
    type = text
  }
  column "competency_model" {
    null    = true
    type    = jsonb
    default = "{}"
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "ACTIVE"
  }
  column "effective_date" {
    null = false
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  primary_key {
    columns = [column.record_id]
  }
  foreign_key "fk_job_roles_family" {
    columns     = [column.parent_record_id, column.tenant_id]
    ref_columns = [table.job_families.column.record_id, table.job_families.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "job_roles_record_id_tenant_id_key" {
    unique  = true
    columns = [column.record_id, column.tenant_id]
  }
  index "job_roles_tenant_id_role_code_effective_date_key" {
    unique  = true
    columns = [column.tenant_id, column.role_code, column.effective_date]
  }
  index "uk_job_roles_current" {
    unique  = true
    columns = [column.tenant_id, column.role_code]
    where   = "(is_current = true)"
  }
  index "uk_job_roles_record" {
    unique  = true
    columns = [column.record_id, column.tenant_id, column.role_code]
  }
}
table "organization_units" {
  schema = schema.public
  column "record_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "code" {
    null = false
    type = character_varying(12)
  }
  column "parent_code" {
    null = true
    type = character_varying(12)
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  column "unit_type" {
    null = false
    type = character_varying(64)
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "ACTIVE"
  }
  column "level" {
    null    = false
    type    = integer
    default = 1
  }
  column "hierarchy_depth" {
    null    = false
    type    = integer
    default = 0
  }
  column "code_path" {
    null    = false
    type    = text
    default = ""
  }
  column "name_path" {
    null    = false
    type    = text
    default = ""
  }
  column "sort_order" {
    null    = false
    type    = integer
    default = 0
  }
  column "description" {
    null = true
    type = text
  }
  column "profile" {
    null    = false
    type    = jsonb
    default = "{}"
  }
  column "metadata" {
    null    = false
    type    = jsonb
    default = "{}"
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "effective_date" {
    null    = false
    type    = date
    default = sql("CURRENT_DATE")
  }
  column "end_date" {
    null = true
    type = date
  }
  column "change_reason" {
    null = true
    type = text
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  column "deleted_by" {
    null = true
    type = uuid
  }
  column "deletion_reason" {
    null = true
    type = text
  }
  column "suspended_at" {
    null = true
    type = timestamptz
  }
  column "suspended_by" {
    null = true
    type = uuid
  }
  column "suspension_reason" {
    null = true
    type = text
  }
  column "operated_by_id" {
    null = true
    type = uuid
  }
  column "operated_by_name" {
    null = true
    type = text
  }
  column "operation_type" {
    null    = true
    type    = character_varying(20)
    default = "CREATE"
  }
  column "effective_from" {
    null = true
    type = timestamptz
  }
  column "effective_to" {
    null = true
    type = timestamptz
  }
  column "changed_by" {
    null = true
    type = uuid
  }
  column "approved_by" {
    null = true
    type = uuid
  }
  primary_key {
    columns = [column.record_id]
  }
  index "idx_org_unit_type_optimized" {
    columns = [column.tenant_id, column.unit_type, column.is_current]
    where   = "(is_current = true)"
  }
  index "idx_org_units_code_current_active" {
    columns = [column.code]
    where   = "((is_current = true) AND ((status)::text <> 'DELETED'::text))"
  }
  index "idx_org_units_parent" {
    columns = [column.parent_code]
  }
  index "idx_org_units_tenant" {
    columns = [column.tenant_id]
  }
  index "idx_organization_current_only" {
    columns = [column.tenant_id, column.code]
    where   = "(is_current = true)"
  }
  index "idx_organization_date_range" {
    columns = [column.tenant_id, column.effective_date, column.end_date]
  }
  index "idx_organization_temporal_main" {
    on {
      column = column.tenant_id
    }
    on {
      column = column.code
    }
    on {
      desc   = true
      column = column.effective_date
    }
    on {
      column = column.is_current
    }
  }
  index "idx_organization_units_effective_from" {
    columns = [column.effective_from]
  }
  index "idx_organization_units_effective_to" {
    columns = [column.effective_to]
  }
  index "ix_org_adjacent_versions" {
    columns = [column.tenant_id, column.code, column.effective_date, column.record_id]
    where   = "((status)::text <> 'DELETED'::text)"
  }
  index "ix_org_current_lookup" {
    columns = [column.tenant_id, column.code, column.is_current]
    where   = "((is_current = true) AND ((status)::text <> 'DELETED'::text))"
  }
  index "ix_org_daily_transition" {
    columns = [column.effective_date, column.end_date, column.is_current]
    where   = "((status)::text <> 'DELETED'::text)"
  }
  index "ix_org_temporal_boundaries" {
    columns = [column.code, column.effective_date, column.end_date, column.is_current]
    where   = "((status)::text <> 'DELETED'::text)"
  }
  index "ix_org_temporal_query" {
    where = "((status)::text <> 'DELETED'::text)"
    on {
      column = column.tenant_id
    }
    on {
      column = column.code
    }
    on {
      desc   = true
      column = column.effective_date
    }
  }
  index "uidx_org_record_id" {
    unique  = true
    columns = [column.record_id]
  }
  index "uk_org_current" {
    unique  = true
    columns = [column.tenant_id, column.code]
    where   = "((is_current = true) AND ((status)::text <> 'DELETED'::text))"
  }
  index "uk_org_current_active_only" {
    unique  = true
    columns = [column.tenant_id, column.code]
    where   = "((is_current = true) AND ((status)::text <> 'DELETED'::text))"
  }
  index "uk_org_temporal_point" {
    unique  = true
    columns = [column.tenant_id, column.code, column.effective_date]
    where   = "((status)::text <> 'DELETED'::text)"
  }
  index "uk_org_ver_active_only" {
    unique  = true
    columns = [column.tenant_id, column.code, column.effective_date]
    where   = "((status)::text <> 'DELETED'::text)"
  }
  check "chk_deleted_not_current" {
    expr = "\nCASE\n    WHEN ((status)::text = 'DELETED'::text) THEN (is_current = false)\n    ELSE true\nEND"
  }
  check "chk_org_units_not_deleted_current" {
    expr = "\nCASE\n    WHEN (((status)::text = 'DELETED'::text) OR (deleted_at IS NOT NULL)) THEN (is_current = false)\n    ELSE true\nEND"
  }
  check "valid_unit_type" {
    expr = "((unit_type)::text = ANY ((ARRAY['DEPARTMENT'::character varying, 'ORGANIZATION_UNIT'::character varying, 'PROJECT_TEAM'::character varying])::text[]))"
  }
}
table "organization_units_backup_temporal" {
  schema = schema.public
  column "record_id" {
    null = true
    type = uuid
  }
  column "tenant_id" {
    null = true
    type = uuid
  }
  column "code" {
    null = true
    type = character_varying(12)
  }
  column "parent_code" {
    null = true
    type = character_varying(12)
  }
  column "name" {
    null = true
    type = character_varying(255)
  }
  column "unit_type" {
    null = true
    type = character_varying(64)
  }
  column "status" {
    null = true
    type = character_varying(20)
  }
  column "level" {
    null = true
    type = integer
  }
  column "hierarchy_depth" {
    null = true
    type = integer
  }
  column "code_path" {
    null = true
    type = text
  }
  column "name_path" {
    null = true
    type = text
  }
  column "sort_order" {
    null = true
    type = integer
  }
  column "description" {
    null = true
    type = text
  }
  column "profile" {
    null = true
    type = jsonb
  }
  column "metadata" {
    null = true
    type = jsonb
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "effective_date" {
    null = true
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "change_reason" {
    null = true
    type = text
  }
  column "is_current" {
    null = true
    type = boolean
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  column "deleted_by" {
    null = true
    type = uuid
  }
  column "deletion_reason" {
    null = true
    type = text
  }
  column "suspended_at" {
    null = true
    type = timestamptz
  }
  column "suspended_by" {
    null = true
    type = uuid
  }
  column "suspension_reason" {
    null = true
    type = text
  }
  column "operated_by_id" {
    null = true
    type = uuid
  }
  column "operated_by_name" {
    null = true
    type = text
  }
  column "operation_type" {
    null = true
    type = character_varying(20)
  }
  column "effective_from" {
    null = true
    type = timestamptz
  }
  column "effective_to" {
    null = true
    type = timestamptz
  }
  column "changed_by" {
    null = true
    type = uuid
  }
  column "approved_by" {
    null = true
    type = uuid
  }
  column "is_temporal" {
    null = true
    type = boolean
  }
}
table "organization_units_unittype_backup" {
  schema = schema.public
  column "record_id" {
    null = true
    type = uuid
  }
  column "tenant_id" {
    null = true
    type = uuid
  }
  column "code" {
    null = true
    type = character_varying(12)
  }
  column "parent_code" {
    null = true
    type = character_varying(12)
  }
  column "name" {
    null = true
    type = character_varying(255)
  }
  column "unit_type" {
    null = true
    type = character_varying(64)
  }
  column "status" {
    null = true
    type = character_varying(20)
  }
  column "level" {
    null = true
    type = integer
  }
  column "hierarchy_depth" {
    null = true
    type = integer
  }
  column "code_path" {
    null = true
    type = text
  }
  column "name_path" {
    null = true
    type = text
  }
  column "sort_order" {
    null = true
    type = integer
  }
  column "description" {
    null = true
    type = text
  }
  column "profile" {
    null = true
    type = jsonb
  }
  column "metadata" {
    null = true
    type = jsonb
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "effective_date" {
    null = true
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "change_reason" {
    null = true
    type = text
  }
  column "is_current" {
    null = true
    type = boolean
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  column "deleted_by" {
    null = true
    type = uuid
  }
  column "deletion_reason" {
    null = true
    type = text
  }
  column "suspended_at" {
    null = true
    type = timestamptz
  }
  column "suspended_by" {
    null = true
    type = uuid
  }
  column "suspension_reason" {
    null = true
    type = text
  }
  column "operated_by_id" {
    null = true
    type = uuid
  }
  column "operated_by_name" {
    null = true
    type = text
  }
  column "operation_type" {
    null = true
    type = character_varying(20)
  }
  column "effective_from" {
    null = true
    type = timestamptz
  }
  column "effective_to" {
    null = true
    type = timestamptz
  }
  column "changed_by" {
    null = true
    type = uuid
  }
  column "approved_by" {
    null = true
    type = uuid
  }
  column "is_temporal" {
    null = true
    type = boolean
  }
}
table "position_assignments" {
  schema = schema.public
  column "assignment_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "position_code" {
    null = false
    type = character_varying(8)
  }
  column "position_record_id" {
    null = false
    type = uuid
  }
  column "employee_id" {
    null = false
    type = uuid
  }
  column "employee_name" {
    null = false
    type = character_varying(255)
  }
  column "employee_number" {
    null = true
    type = character_varying(64)
  }
  column "assignment_type" {
    null = false
    type = character_varying(20)
  }
  column "assignment_status" {
    null    = false
    type    = character_varying(20)
    default = "ACTIVE"
  }
  column "fte" {
    null    = false
    type    = numeric(5,2)
    default = 1
  }
  column "effective_date" {
    null = false
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "notes" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "acting_until" {
    null = true
    type = date
  }
  column "auto_revert" {
    null    = false
    type    = boolean
    default = false
  }
  column "reminder_sent_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.assignment_id]
  }
  foreign_key "fk_position_assignments_position" {
    columns     = [column.tenant_id, column.position_code, column.position_record_id]
    ref_columns = [table.positions.column.tenant_id, table.positions.column.code, table.positions.column.record_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_position_assignments_auto_revert_due" {
    columns = [column.tenant_id, column.auto_revert, column.acting_until]
    where   = "((assignment_type)::text = 'ACTING'::text)"
  }
  index "idx_position_assignments_employee" {
    on {
      column = column.tenant_id
    }
    on {
      column = column.employee_id
    }
    on {
      desc   = true
      column = column.effective_date
    }
  }
  index "idx_position_assignments_position" {
    on {
      column = column.tenant_id
    }
    on {
      column = column.position_code
    }
    on {
      desc   = true
      column = column.effective_date
    }
  }
  index "idx_position_assignments_status" {
    columns = [column.tenant_id, column.assignment_status, column.is_current]
  }
  index "uk_position_assignments_active" {
    unique  = true
    columns = [column.tenant_id, column.position_code, column.employee_id]
    where   = "((is_current = true) AND ((assignment_status)::text = 'ACTIVE'::text))"
  }
  index "uk_position_assignments_effective" {
    unique  = true
    columns = [column.tenant_id, column.position_code, column.employee_id, column.effective_date]
  }
  check "chk_position_assignments_auto_revert" {
    expr = "((auto_revert = false) OR (((assignment_type)::text = 'ACTING'::text) AND (acting_until IS NOT NULL)))"
  }
  check "chk_position_assignments_dates" {
    expr = "(((end_date IS NULL) OR (end_date > effective_date)) AND ((acting_until IS NULL) OR (acting_until > effective_date)))"
  }
  check "chk_position_assignments_fte" {
    expr = "((fte >= (0)::numeric) AND (fte <= (1)::numeric))"
  }
  check "chk_position_assignments_status" {
    expr = "((assignment_status)::text = ANY ((ARRAY['PENDING'::character varying, 'ACTIVE'::character varying, 'ENDED'::character varying])::text[]))"
  }
  check "chk_position_assignments_type" {
    expr = "((assignment_type)::text = ANY ((ARRAY['PRIMARY'::character varying, 'SECONDARY'::character varying, 'ACTING'::character varying])::text[]))"
  }
}
table "positions" {
  schema = schema.public
  column "record_id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = uuid
  }
  column "code" {
    null = false
    type = character_varying(8)
  }
  column "title" {
    null = false
    type = character_varying(120)
  }
  column "job_profile_code" {
    null = true
    type = character_varying(64)
  }
  column "job_profile_name" {
    null = true
    type = character_varying(255)
  }
  column "job_family_group_code" {
    null = false
    type = character_varying(20)
  }
  column "job_family_group_name" {
    null = false
    type = character_varying(255)
  }
  column "job_family_group_record_id" {
    null = false
    type = uuid
  }
  column "job_family_code" {
    null = false
    type = character_varying(20)
  }
  column "job_family_name" {
    null = false
    type = character_varying(255)
  }
  column "job_family_record_id" {
    null = false
    type = uuid
  }
  column "job_role_code" {
    null = false
    type = character_varying(20)
  }
  column "job_role_name" {
    null = false
    type = character_varying(255)
  }
  column "job_role_record_id" {
    null = false
    type = uuid
  }
  column "job_level_code" {
    null = false
    type = character_varying(20)
  }
  column "job_level_name" {
    null = false
    type = character_varying(255)
  }
  column "job_level_record_id" {
    null = false
    type = uuid
  }
  column "organization_code" {
    null = false
    type = character_varying(7)
  }
  column "organization_name" {
    null = true
    type = character_varying(255)
  }
  column "position_type" {
    null = false
    type = character_varying(50)
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "PLANNED"
  }
  column "employment_type" {
    null = false
    type = character_varying(50)
  }
  column "headcount_capacity" {
    null    = false
    type    = numeric(5,2)
    default = 1
  }
  column "headcount_in_use" {
    null    = false
    type    = numeric(5,2)
    default = 0
  }
  column "grade_level" {
    null = true
    type = character_varying(20)
  }
  column "cost_center_code" {
    null = true
    type = character_varying(50)
  }
  column "reports_to_position_code" {
    null = true
    type = character_varying(8)
  }
  column "profile" {
    null    = false
    type    = jsonb
    default = "{}"
  }
  column "effective_date" {
    null = false
    type = date
  }
  column "end_date" {
    null = true
    type = date
  }
  column "is_current" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  column "operation_type" {
    null    = false
    type    = character_varying(20)
    default = "CREATE"
  }
  column "operated_by_id" {
    null = false
    type = uuid
  }
  column "operated_by_name" {
    null = false
    type = character_varying(255)
  }
  column "operation_reason" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.record_id]
  }
  foreign_key "fk_positions_family" {
    columns     = [column.job_family_record_id, column.tenant_id]
    ref_columns = [table.job_families.column.record_id, table.job_families.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_positions_family_group" {
    columns     = [column.job_family_group_record_id, column.tenant_id]
    ref_columns = [table.job_family_groups.column.record_id, table.job_family_groups.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_positions_level" {
    columns     = [column.job_level_record_id, column.tenant_id]
    ref_columns = [table.job_levels.column.record_id, table.job_levels.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_positions_role" {
    columns     = [column.job_role_record_id, column.tenant_id]
    ref_columns = [table.job_roles.column.record_id, table.job_roles.column.tenant_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_positions_current" {
    columns = [column.tenant_id]
    where   = "(is_current = true)"
  }
  index "idx_positions_effective_date" {
    columns = [column.tenant_id, column.effective_date]
  }
  index "idx_positions_job_family" {
    columns = [column.tenant_id, column.job_family_code, column.is_current]
  }
  index "idx_positions_job_family_group" {
    columns = [column.tenant_id, column.job_family_group_code, column.is_current]
  }
  index "idx_positions_job_role" {
    columns = [column.tenant_id, column.job_role_code, column.is_current]
  }
  index "idx_positions_org_code" {
    columns = [column.tenant_id, column.organization_code, column.is_current]
  }
  index "idx_positions_status" {
    columns = [column.tenant_id, column.status, column.is_current]
  }
  index "positions_tenant_id_code_effective_date_key" {
    unique  = true
    columns = [column.tenant_id, column.code, column.effective_date]
  }
  index "positions_tenant_id_code_record_id_key" {
    unique  = true
    columns = [column.tenant_id, column.code, column.record_id]
  }
  index "uk_positions_current_active" {
    unique  = true
    columns = [column.tenant_id, column.code]
    where   = "((is_current = true) AND ((status)::text <> 'DELETED'::text))"
  }
}
schema "public" {
  comment = "standard public schema"
}
