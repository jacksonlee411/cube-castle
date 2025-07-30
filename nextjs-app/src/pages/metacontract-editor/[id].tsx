import React from 'react';
import { useRouter } from 'next/router';
import { MetaContractEditor } from '@/components/metacontract-editor/MetaContractEditor';

const ProjectEditorPage = () => {
  const router = useRouter();
  const { id } = router.query;

  if (!id || typeof id !== 'string') {
    return (
      <div className="flex items-center justify-center h-screen">
        <p>Invalid project ID</p>
      </div>
    );
  }

  return (
    <div className="h-screen">
      <MetaContractEditor
        projectId={id}
        readonly={false}
      />
    </div>
  );
};

export default ProjectEditorPage;