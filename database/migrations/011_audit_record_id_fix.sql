-- 011_audit_record_id_fix.sql
-- 修复审计历史查询问题：添加record_id字段到audit_logs表
-- 解决前端按recordId查询审计历史时找不到记录的问题

-- Migration: 添加record_id字段到audit_logs表
-- Author: System
-- Date: 2025-09-06
-- Related Issue: 审计历史页签查询不到修改记录

-- 第一步：添加record_id列（允许NULL，稍后回填）
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS record_id UUID;

COMMENT ON COLUMN audit_logs.record_id IS '组织单元时态版本的唯一标识，用于精确审计查询';
