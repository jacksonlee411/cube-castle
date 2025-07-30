import React from 'react';
import { useRouter } from 'next/router';
import { MetaContractEditor } from '@/components/metacontract-editor/MetaContractEditor';

const MetaContractEditorPage = () => {
  const router = useRouter();
  const { projectId } = router.query;

  const defaultContent = `specification_version: "1.0"
api_id: "550e8400-e29b-41d4-a716-446655440000"
namespace: "corehr"
resource_name: "employee"
version: "1.0.0"

data_structure:
  primary_key: "id"
  data_classification: "pii"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "public"
    - name: "first_name"
      type: "string"
      required: true
      data_classification: "pii"
    - name: "last_name"
      type: "string"
      required: true
      data_classification: "pii"
    - name: "email"
      type: "string"
      required: true
      unique: true
      data_classification: "pii"
    - name: "employee_id"
      type: "string"
      required: true
      unique: true
      data_classification: "internal"

security_model:
  tenant_isolation: true
  access_control: "rbac"
  data_classification: "pii"
  compliance_tags: ["gdpr", "ccpa"]

temporal_behavior:
  temporality_paradigm: "event_sourced"
  state_transition_model: "status_based"
  history_retention: "indefinite"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: true
  events_enabled: true

relationships: []`;

  return (
    <div className="h-screen">
      <MetaContractEditor
        projectId={projectId as string}
        initialContent={projectId ? undefined : defaultContent}
        readonly={false}
      />
    </div>
  );
};

export default MetaContractEditorPage;