import { useState, useCallback, useEffect } from 'react';

interface EditorProject {
  id: string;
  name: string;
  description: string;
  content: string;
  version: string;
  status: 'draft' | 'compiling' | 'valid' | 'error' | 'published';
  created_at: string;
  updated_at: string;
  last_compiled?: string;
  compile_error?: string;
}

interface CompileResults {
  success: boolean;
  errors?: any[];
  warnings?: any[];
  generated_files?: Record<string, string>;
  schema?: any;
  compile_time?: string;
}

interface CreateProjectRequest {
  name: string;
  description: string;
  content: string;
}

interface UpdateProjectRequest {
  name?: string;
  description?: string;
  content?: string;
}

export const useMetaContractEditor = (projectId?: string) => {
  const [project, setProject] = useState<EditorProject | null>(null);
  const [projects, setProjects] = useState<EditorProject[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [compileResults, setCompileResults] = useState<CompileResults | null>(null);

  // API base URL
  const API_BASE = '/api/v1/metacontract-editor';

  // Generic API call helper
  const apiCall = useCallback(async (endpoint: string, options: RequestInit = {}) => {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      credentials: 'include', // Include cookies for authentication
      ...options,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP ${response.status}`);
    }

    return response.json();
  }, []);

  // Load a specific project
  const loadProject = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const data = await apiCall(`/projects/${id}`);
      setProject(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load project');
    } finally {
      setIsLoading(false);
    }
  }, [apiCall]);

  // Load list of projects
  const loadProjects = useCallback(async (limit = 20, offset = 0) => {
    setIsLoading(true);
    setError(null);

    try {
      const data = await apiCall(`/projects?limit=${limit}&offset=${offset}`);
      setProjects(data.projects || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setIsLoading(false);
    }
  }, [apiCall]);

  // Create a new project
  const createProject = useCallback(async (projectData: CreateProjectRequest) => {
    setIsLoading(true);
    setError(null);

    try {
      const data = await apiCall('/projects', {
        method: 'POST',
        body: JSON.stringify(projectData),
      });
      
      setProject(data);
      return data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [apiCall]);

  // Update an existing project
  const updateProject = useCallback(async (id: string, updates: UpdateProjectRequest) => {
    setIsLoading(true);
    setError(null);

    try {
      const data = await apiCall(`/projects/${id}`, {
        method: 'PUT',
        body: JSON.stringify(updates),
      });
      
      setProject(data);
      return data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update project');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [apiCall]);

  // Save project (wrapper for updateProject)
  const saveProject = useCallback(async (id: string, updates: UpdateProjectRequest) => {
    return updateProject(id, updates);
  }, [updateProject]);

  // Delete a project
  const deleteProject = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);

    try {
      await apiCall(`/projects/${id}`, {
        method: 'DELETE',
      });
      
      // Remove from local state
      setProjects(prev => prev.filter(p => p.id !== id));
      
      if (project?.id === id) {
        setProject(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete project');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [apiCall, project]);

  // Compile a project
  const compileProject = useCallback(async (id: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const data = await apiCall(`/projects/${id}/compile`, {
        method: 'POST',
      });
      
      setCompileResults(data);
      
      // Update project status if it was returned
      if (project?.id === id) {
        setProject(prev => prev ? {
          ...prev,
          status: data.success ? 'valid' : 'error',
          last_compiled: new Date().toISOString(),
          compile_error: data.success ? undefined : JSON.stringify(data.errors)
        } : null);
      }
      
      return data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to compile project');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [apiCall, project]);

  // Compile content (preview mode)
  const compileContent = useCallback(async (content: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const data = await apiCall('/compile', {
        method: 'POST',
        body: JSON.stringify({
          content,
          preview: true
        }),
      });
      
      setCompileResults(data);
      return data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to compile content');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [apiCall]);

  // Start an editing session
  const startSession = useCallback(async (projectId: string) => {
    try {
      const data = await apiCall(`/projects/${projectId}/sessions`, {
        method: 'POST',
      });
      return data;
    } catch (err) {
      console.error('Failed to start session:', err);
      throw err;
    }
  }, [apiCall]);

  // End an editing session
  const endSession = useCallback(async (sessionId: string) => {
    try {
      await apiCall(`/sessions/${sessionId}`, {
        method: 'DELETE',
      });
    } catch (err) {
      console.error('Failed to end session:', err);
      throw err;
    }
  }, [apiCall]);

  // Get available templates
  const getTemplates = useCallback(async (category?: string) => {
    try {
      const query = category ? `?category=${encodeURIComponent(category)}` : '';
      const data = await apiCall(`/templates${query}`);
      return data.templates || [];
    } catch (err) {
      console.error('Failed to get templates:', err);
      throw err;
    }
  }, [apiCall]);

  // Auto-load project if projectId is provided
  useEffect(() => {
    if (projectId && !project) {
      loadProject(projectId);
    }
  }, [projectId, project, loadProject]);

  return {
    // State
    project,
    projects,
    isLoading,
    error,
    compileResults,

    // Project operations
    loadProject,
    loadProjects,
    createProject,
    updateProject,
    saveProject,
    deleteProject,

    // Compilation
    compileProject,
    compileContent,

    // Sessions
    startSession,
    endSession,

    // Templates
    getTemplates,

    // Utilities
    clearError: () => setError(null),
    clearResults: () => setCompileResults(null)
  };
};