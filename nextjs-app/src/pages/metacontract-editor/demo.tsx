// src/pages/metacontract-editor/demo.tsx - Meta-Contract编辑器演示页面
import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';

interface Project {
  id: string;
  name: string;
  description: string;
  content: string;
  version: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface Template {
  id: string;
  name: string;
  description: string;
  category: string;
  content: string;
  tags: string[];
}

const MetaContractEditorDemo: React.FC = () => {
  const router = useRouter();
  const [projects, setProjects] = useState<Project[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [content, setContent] = useState('');
  const [projectName, setProjectName] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    loadProjects();
    loadTemplates();
  }, []);

  const loadProjects = async () => {
    try {
      const response = await fetch('/api/v1/metacontract/projects');
      if (response.ok) {
        const data = await response.json();
        setProjects(data.projects || []);
      }
    } catch (error) {
      console.error('Failed to load projects:', error);
    }
  };

  const loadTemplates = async () => {
    try {
      const response = await fetch('/api/v1/metacontract/templates');
      if (response.ok) {
        const data = await response.json();
        setTemplates(data.templates || []);
      }
    } catch (error) {
      console.error('Failed to load templates:', error);
    }
  };

  const handleTemplateSelect = (template: Template) => {
    setSelectedTemplate(template);
    setContent(template.content);
    setMessage(`已加载模板: ${template.name}`);
  };

  const handleCreateProject = async () => {
    if (!projectName.trim()) {
      setMessage('请输入项目名称');
      return;
    }

    setLoading(true);
    try {
      const response = await fetch('/api/v1/metacontract/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: projectName,
          description: `基于模板创建: ${selectedTemplate?.name || '自定义内容'}`,
          content: content,
        }),
      });

      if (response.ok) {
        const newProject = await response.json();
        setProjects([newProject, ...projects]);
        setMessage(`项目 "${projectName}" 创建成功！`);
        setProjectName('');
      } else {
        setMessage('创建项目失败');
      }
    } catch (error) {
      setMessage('创建项目时发生错误');
      console.error('Error creating project:', error);
    } finally {
      setLoading(false);
    }
  };

  const compileProject = async (projectId: string) => {
    setLoading(true);
    try {
      const response = await fetch(`/api/v1/metacontract/projects/${projectId}/compile`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          content: content,
          preview: true,
        }),
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success) {
          setMessage('编译成功！✅');
        } else {
          setMessage(`编译失败: ${result.errors?.[0]?.message || '未知错误'}`);
        }
      }
    } catch (error) {
      setMessage('编译时发生错误');
      console.error('Error compiling project:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ 
      fontFamily: 'Arial, sans-serif', 
      maxWidth: '1400px', 
      margin: '0 auto', 
      padding: '20px',
      lineHeight: '1.6'
    }}>
      <header style={{ 
        textAlign: 'center', 
        marginBottom: '30px',
        borderBottom: '2px solid #722ed1',
        paddingBottom: '20px'
      }}>
        <h1 style={{ 
          color: '#722ed1', 
          fontSize: '2.2rem',
          margin: '0 0 10px 0'
        }}>
          📝 Meta-Contract编辑器演示
        </h1>
        <p style={{ 
          color: '#666', 
          fontSize: '1.1rem',
          margin: '0'
        }}>
          智能化的元合约编辑器，支持YAML语法、实时编译和模板管理
        </p>
        <button
          onClick={() => router.back()}
          style={{
            marginTop: '10px',
            padding: '8px 16px',
            backgroundColor: '#f0f0f0',
            border: '1px solid #d9d9d9',
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          ← 返回主页
        </button>
      </header>

      {message && (
        <div style={{ 
          backgroundColor: message.includes('失败') || message.includes('错误') ? '#fff2f0' : '#f6ffed',
          border: `1px solid ${message.includes('失败') || message.includes('错误') ? '#ffccc7' : '#b7eb8f'}`,
          borderRadius: '4px',
          padding: '10px',
          marginBottom: '20px',
          color: message.includes('失败') || message.includes('错误') ? '#cf1322' : '#389e0d'
        }}>
          {message}
        </div>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '20px' }}>
        {/* 左侧面板：模板和项目 */}
        <div>
          {/* 模板选择 */}
          <div style={{ 
            border: '1px solid #d9d9d9', 
            borderRadius: '8px', 
            padding: '15px',
            marginBottom: '20px'
          }}>
            <h3 style={{ margin: '0 0 15px 0', color: '#722ed1' }}>📋 选择模板</h3>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {templates.map((template) => (
                <button
                  key={template.id}
                  onClick={() => handleTemplateSelect(template)}
                  style={{
                    padding: '10px',
                    textAlign: 'left',
                    border: '1px solid #d9d9d9',
                    borderRadius: '4px',
                    backgroundColor: selectedTemplate?.id === template.id ? '#f0f5ff' : 'white',
                    cursor: 'pointer',
                    transition: 'background-color 0.2s'
                  }}
                  onMouseEnter={(e) => {
                    if (selectedTemplate?.id !== template.id) {
                      e.currentTarget.style.backgroundColor = '#fafafa';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (selectedTemplate?.id !== template.id) {
                      e.currentTarget.style.backgroundColor = 'white';
                    }
                  }}
                >
                  <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>
                    {template.name}
                  </div>
                  <div style={{ fontSize: '0.9rem', color: '#666' }}>
                    {template.description}
                  </div>
                  <div style={{ fontSize: '0.8rem', color: '#999', marginTop: '4px' }}>
                    {template.tags.join(', ')}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* 项目列表 */}
          <div style={{ 
            border: '1px solid #d9d9d9', 
            borderRadius: '8px', 
            padding: '15px'
          }}>
            <h3 style={{ margin: '0 0 15px 0', color: '#722ed1' }}>💾 现有项目</h3>
            <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
              {projects.length === 0 ? (
                <p style={{ color: '#999', textAlign: 'center', margin: '20px 0' }}>
                  暂无项目，创建您的第一个项目！
                </p>
              ) : (
                projects.map((project) => (
                  <div
                    key={project.id}
                    style={{
                      padding: '10px',
                      border: '1px solid #f0f0f0',
                      borderRadius: '4px',
                      marginBottom: '8px',
                      backgroundColor: '#fafafa'
                    }}
                  >
                    <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>
                      {project.name}
                    </div>
                    <div style={{ fontSize: '0.9rem', color: '#666', marginBottom: '8px' }}>
                      {project.description}
                    </div>
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                      <span style={{ 
                        fontSize: '0.8rem', 
                        padding: '2px 6px', 
                        backgroundColor: project.status === 'draft' ? '#faad14' : '#52c41a',
                        color: 'white',
                        borderRadius: '3px'
                      }}>
                        {project.status}
                      </span>
                      <button
                        onClick={() => compileProject(project.id)}
                        disabled={loading}
                        style={{
                          fontSize: '0.8rem',
                          padding: '4px 8px',
                          backgroundColor: '#722ed1',
                          color: 'white',
                          border: 'none',
                          borderRadius: '3px',
                          cursor: loading ? 'not-allowed' : 'pointer',
                          opacity: loading ? 0.6 : 1
                        }}
                      >
                        编译
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* 右侧面板：编辑器 */}
        <div>
          <div style={{ 
            border: '1px solid #d9d9d9', 
            borderRadius: '8px', 
            padding: '15px'
          }}>
            <h3 style={{ margin: '0 0 15px 0', color: '#722ed1' }}>✏️ 内容编辑</h3>
            
            {/* 项目创建 */}
            <div style={{ marginBottom: '20px', display: 'flex', gap: '10px', alignItems: 'center' }}>
              <input
                type="text"
                placeholder="输入项目名称..."
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                style={{
                  flex: 1,
                  padding: '8px 12px',
                  border: '1px solid #d9d9d9',
                  borderRadius: '4px',
                  fontSize: '14px'
                }}
              />
              <button
                onClick={handleCreateProject}
                disabled={loading || !projectName.trim()}
                style={{
                  padding: '8px 16px',
                  backgroundColor: '#722ed1',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: loading || !projectName.trim() ? 'not-allowed' : 'pointer',
                  opacity: loading || !projectName.trim() ? 0.6 : 1
                }}
              >
                {loading ? '创建中...' : '创建项目'}
              </button>
            </div>

            {/* 内容编辑器 */}
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="在此输入或选择模板开始编辑..."
              style={{
                width: '100%',
                height: '500px',
                padding: '12px',
                border: '1px solid #d9d9d9',
                borderRadius: '4px',
                fontSize: '14px',
                fontFamily: 'Monaco, Consolas, "Courier New", monospace',
                resize: 'vertical',
                lineHeight: '1.5'
              }}
            />

            <div style={{ 
              marginTop: '15px', 
              padding: '10px', 
              backgroundColor: '#f8f9fa', 
              borderRadius: '4px',
              fontSize: '0.9rem',
              color: '#666'
            }}>
              <strong>功能说明：</strong>
              <ul style={{ margin: '8px 0', paddingLeft: '20px' }}>
                <li>选择左侧模板快速开始</li>
                <li>在编辑器中修改YAML内容</li>
                <li>输入项目名称并点击"创建项目"</li>
                <li>点击现有项目的"编译"按钮测试编译</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div style={{ 
        marginTop: '30px',
        textAlign: 'center',
        padding: '20px',
        backgroundColor: '#f0f8ff',
        border: '1px solid #1890ff',
        borderRadius: '8px'
      }}>
        <h4 style={{ color: '#1890ff', margin: '0 0 10px 0' }}>🚀 API状态</h4>
        <div style={{ display: 'flex', justifyContent: 'center', gap: '20px', flexWrap: 'wrap' }}>
          <span style={{ color: '#52c41a' }}>✅ 后端API (http://localhost:8080)</span>
          <span style={{ color: '#52c41a' }}>✅ 项目管理接口</span>
          <span style={{ color: '#52c41a' }}>✅ 模板库接口</span>
          <span style={{ color: '#52c41a' }}>✅ 编译接口</span>
        </div>
      </div>
    </div>
  );
};

export default MetaContractEditorDemo;