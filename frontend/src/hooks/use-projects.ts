// 📦 Project Management Hook

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useApi } from '@/contexts/api-context';
import type { Project } from '@/lib/api';

// 🎣 Hook for managing projects
export function useProjects() {
  const { client, isLoading: isClientLoading } = useApi();
  const queryClient = useQueryClient();

  // 📋 List projects
  const {
    data: projects,
    isLoading: isLoadingProjects,
    error: projectsError,
  } = useQuery({
    queryKey: ['projects'],
    queryFn: () => client?.listProjects() ?? Promise.resolve([]),
    enabled: !!client && !isClientLoading,
  });

  // ➕ Create project
  const createProjectMutation = useMutation({
    mutationFn: (params: { name: string; description: string }) =>
      client?.createProject(params.name, params.description) ?? Promise.reject('No client available'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });

  // 🔍 Get single project
  const useProject = (projectId: string) =>
    useQuery({
      queryKey: ['projects', projectId],
      queryFn: () => client?.getProject(projectId) ?? Promise.reject('No client available'),
      enabled: !!client && !isClientLoading && !!projectId,
    });

  return {
    // 📊 Data
    projects,
    isLoadingProjects,
    projectsError,

    // 🛠️ Operations
    createProject: createProjectMutation.mutate,
    isCreatingProject: createProjectMutation.isPending,
    createProjectError: createProjectMutation.error,

    // 🎣 Sub-hooks
    useProject,
  };
} 