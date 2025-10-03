import { logger } from '@/shared/utils/logger';
import { useCallback, useEffect, useMemo, useState } from "react";
import { normalizeParentCode } from "@/shared/utils/organization-helpers";
import type { OrganizationStateMutationResult } from "@/shared/hooks/useOrganizationMutations";
import type { OrganizationRequest } from "@/shared/types/organization";
import type { TemporalVersionPayload } from "@/shared/types/temporal";
import type { TemporalEditFormData } from "../TemporalEditForm";
import type { TimelineVersion } from "../TimelineComponent";
import type { TabType } from "../TabNavigation";
import {
  fetchOrganizationVersions,
  fetchHierarchyPaths,
  deactivateOrganizationVersion,
  createOrganizationUnit,
  createTemporalVersion,
  updateHistoryRecord,
  type HierarchyPaths,
} from "./temporalMasterDetailApi";

type ApiOrganizationStatus = "ACTIVE" | "INACTIVE" | "PLANNED";

const mapLifecycleStatusToApiStatus = (
  lifecycleStatus: TemporalEditFormData["lifecycleStatus"],
): ApiOrganizationStatus => {
  switch (lifecycleStatus) {
    case "CURRENT":
      return "ACTIVE";
    case "PLANNED":
      return "PLANNED";
    case "HISTORICAL":
    case "INACTIVE":
    case "DELETED":
    default:
      return "INACTIVE";
  }
};

export interface TemporalMasterDetailViewProps {
  organizationCode: string | null;
  readonly?: boolean;
  onBack?: () => void;
  onCreateSuccess?: (newOrganizationCode: string) => void;
  isCreateMode?: boolean;
}

interface FormInitialData {
  name: string;
  unitType: string;
  status: string;
  lifecycleStatus?: TimelineVersion["lifecycleStatus"];
  description?: string;
  parentCode?: string;
  effectiveDate?: string;
}

export interface TemporalMasterDetailState {
  versions: TimelineVersion[];
  selectedVersion: TimelineVersion | null;
  isLoading: boolean;
  showDeleteConfirm: TimelineVersion | null;
  isDeleting: boolean;
  loadingError: string | null;
  successMessage: string | null;
  error: string | null;
  retryCount: number;
  isSubmitting: boolean;
  currentETag: string | null;
  activeTab: TabType;
  formMode: "create" | "edit";
  formInitialData: FormInitialData | null;
  displayPaths: HierarchyPaths | null;
  currentTimelineStatus: string | undefined;
  currentOrganizationName: string;
}

export interface TemporalMasterDetailHandlers {
  setShowDeleteConfirm: (version: TimelineVersion | null) => void;
  loadVersions: (isRetry?: boolean, focusRecordId?: string) => Promise<void>;
  handleStateMutationCompleted: (
    action: "suspend" | "activate",
    result: OrganizationStateMutationResult,
  ) => Promise<void>;
  handleDeleteVersion: (version: TimelineVersion) => Promise<void>;
  handleVersionSelect: (version: TimelineVersion) => void;
  handleFormSubmit: (formData: TemporalEditFormData) => Promise<void>;
  handleHistoryEditClose: () => void;
  handleHistoryEditSubmit: (versionData: TemporalVersionPayload & { recordId: string }) => Promise<void>;
  setActiveTab: (tab: TabType) => void;
  setCurrentETag: (etag: string | null) => void;
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
}

export type UseTemporalMasterDetailResult = [
  TemporalMasterDetailState,
  TemporalMasterDetailHandlers,
];

export const useTemporalMasterDetail = (
  options: TemporalMasterDetailViewProps,
): UseTemporalMasterDetailResult => {
  const {
    organizationCode,
    onBack,
    onCreateSuccess,
    isCreateMode = false,
  } = options;

  const [versions, setVersions] = useState<TimelineVersion[]>([]);
  const [selectedVersion, setSelectedVersion] =
    useState<TimelineVersion | null>(null);
  const [isLoading, setIsLoading] = useState(() => Boolean(organizationCode));
  const [showDeleteConfirm, setShowDeleteConfirm] =
    useState<TimelineVersion | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [loadingError, setLoadingError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentETag, setCurrentETag] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>("edit-history");
  const [formMode, setFormMode] = useState<"create" | "edit">(() =>
    isCreateMode ? "create" : "edit",
  );
  const [formInitialData, setFormInitialData] = useState<FormInitialData | null>(
    null,
  );
  const [displayPaths, setDisplayPaths] = useState<HierarchyPaths | null>(null);

  const notifySuccess = useCallback((message: string) => {
    setError(null);
    setSuccessMessage(message);
    setTimeout(() => setSuccessMessage(null), 3000);
  }, []);

  const notifyError = useCallback((message: string) => {
    setSuccessMessage(null);
    setError(message);
    setTimeout(() => setError(null), 5000);
  }, []);

  const loadVersions = useCallback(
    async (isRetry = false, focusRecordId?: string) => {
      if (!organizationCode) {
        setIsLoading(false);
        setVersions([]);
        setSelectedVersion(null);
        setCurrentETag(null);
        return;
      }

      try {
        setIsLoading(true);
        setLoadingError(null);
        if (!isRetry) {
          setRetryCount(0);
        }
        const { versions: loadedVersions, fallbackMessage } =
          await fetchOrganizationVersions(organizationCode);

        setVersions(loadedVersions);

        if (fallbackMessage) {
          setLoadingError(fallbackMessage);
          setTimeout(() => setLoadingError(null), 3000);
        } else if (isRetry) {
          setSuccessMessage("数据加载成功！");
          setTimeout(() => setSuccessMessage(null), 3000);
        }

        const preferredVersion = focusRecordId
          ? loadedVersions.find((v) => v.recordId === focusRecordId) || null
          : null;

        const currentVersion =
          preferredVersion ||
          loadedVersions.find((v) => v.isCurrent) ||
          loadedVersions.at(-1) ||
          null;

        const nextETag =
          loadedVersions.find((v) => v.isCurrent)?.recordId ?? null;
        setCurrentETag(nextETag);

        if (currentVersion) {
          setSelectedVersion(currentVersion);
          setFormMode("edit");
          setFormInitialData({
            name: currentVersion.name,
            unitType: currentVersion.unitType,
            status: currentVersion.status,
            lifecycleStatus: currentVersion.lifecycleStatus,
            description: currentVersion.description || "",
            parentCode: normalizeParentCode.forForm(currentVersion.parentCode),
            effectiveDate: currentVersion.effectiveDate,
          });

          try {
            const hierarchy = await fetchHierarchyPaths(currentVersion.code);
            setDisplayPaths(hierarchy);
          } catch (pathError) {
            logger.warn("加载组织层级路径失败（忽略，不阻塞详情展示）:", pathError);
            setDisplayPaths(null);
          }
        } else {
          setSelectedVersion(null);
          setFormInitialData(null);
          setDisplayPaths(null);
        }
      } catch (loadingException) {
        logger.error("Error loading temporal versions:", loadingException);
        const errorMessage =
          loadingException instanceof Error
            ? loadingException.message
            : "加载版本数据时发生未知错误";
        setLoadingError(errorMessage);
        setRetryCount((prev) => prev + 1);
      } finally {
        setIsLoading(false);
      }
    },
    [organizationCode],
  );

  const handleStateMutationCompleted = useCallback(
    async (
      _action: "suspend" | "activate",
      result: OrganizationStateMutationResult,
    ) => {
      if (result?.etag) {
        setCurrentETag(result.etag);
      }
      try {
        await loadVersions();
      } catch (mutationRefreshError) {
        logger.warn("状态变更后刷新失败:", mutationRefreshError);
      }
    },
    [loadVersions],
  );

  const handleDeleteVersion = useCallback(
    async (version: TimelineVersion) => {
      if (!version || isDeleting) {
        return;
      }

      try {
        setIsDeleting(true);
        const timeline = await deactivateOrganizationVersion(
          organizationCode,
          version,
        );

        if (timeline) {
          setVersions(timeline);
          const current = timeline.find((v) => v.isCurrent) || timeline[0] || null;
          setSelectedVersion(current);
          setCurrentETag(current?.recordId ?? null);
        } else {
          try {
            await loadVersions();
          } catch (refreshError) {
            logger.warn("数据刷新失败，但删除操作已成功:", refreshError);
          }
        }

        setShowDeleteConfirm(null);
        if (
          !timeline &&
          selectedVersion?.effectiveDate === version.effectiveDate
        ) {
          setSelectedVersion(null);
        }
      } catch (error) {
        logger.error("Error deactivating version:", error);
        throw error;
      } finally {
        setIsDeleting(false);
      }
    },
    [organizationCode, selectedVersion, isDeleting, loadVersions],
  );

  const handleVersionSelect = useCallback(
    (version: TimelineVersion) => {
      setSelectedVersion(version);
      setFormMode("edit");
    setFormInitialData({
      name: version.name,
      unitType: version.unitType,
      status: version.status,
      lifecycleStatus: version.lifecycleStatus,
      description: version.description || "",
      parentCode: normalizeParentCode.forForm(version.parentCode),
      effectiveDate: version.effectiveDate,
    });
    setActiveTab("edit-history");

    (async () => {
      try {
        const hierarchy = await fetchHierarchyPaths(version.code);
        setDisplayPaths(hierarchy);
      } catch (pathError) {
        logger.warn("加载组织层级路径失败（忽略，不阻塞详情展示）:", pathError);
        setDisplayPaths(null);
      }
    })();
    },
    [],
  );

  const handleFormSubmit = useCallback(
    async (formData: TemporalEditFormData) => {
      setIsSubmitting(true);
      try {
        const mappedStatus = mapLifecycleStatusToApiStatus(
          formData.lifecycleStatus,
        );
        if (isCreateMode) {
          const requestBody: OrganizationRequest = {
            name: formData.name,
            unitType: formData.unitType as OrganizationRequest["unitType"],
            description: formData.description || "",
            parentCode: normalizeParentCode.forAPI(formData.parentCode),
            effectiveDate: formData.effectiveDate,
            status: mappedStatus,
            operationReason: formData.changeReason,
          };

          logger.info("提交创建组织请求:", requestBody);

          const newOrganizationCode = await createOrganizationUnit(requestBody);

          if (newOrganizationCode && onCreateSuccess) {
            onCreateSuccess(newOrganizationCode);
            return;
          }

          logger.error("创建成功但未返回组织编码");
          notifyError("创建成功，但未能获取新组织编码，请手动刷新页面");
        } else {
          const versionPayload: TemporalVersionPayload = {
            name: formData.name,
            unitType: formData.unitType as TemporalVersionPayload["unitType"],
            parentCode: normalizeParentCode.forAPI(formData.parentCode),
            description: formData.description || null,
            effectiveDate: formData.effectiveDate,
            status: mappedStatus,
            lifecycleStatus: formData.lifecycleStatus,
          };

          const trimmedReason = formData.changeReason?.trim();
          if (trimmedReason) {
            versionPayload.operationReason = trimmedReason;
          }

          await createTemporalVersion(organizationCode, versionPayload);

          await loadVersions();
          setActiveTab("edit-history");
          setFormMode("create");
          setFormInitialData(null);
          notifySuccess("时态版本创建成功！");
        }
      } catch (submissionError) {
        logger.error(
          isCreateMode ? "创建组织失败:" : "创建时态版本失败:",
          submissionError,
        );

        let errorMessage = isCreateMode ? "创建组织失败" : "创建时态版本失败";
        if (submissionError instanceof Error) {
          errorMessage = submissionError.message;
        } else if (submissionError && typeof submissionError === "string") {
          errorMessage = submissionError;
        }

        notifyError(errorMessage);
      } finally {
        setIsSubmitting(false);
      }
    },
    [
      organizationCode,
      loadVersions,
      isCreateMode,
      onCreateSuccess,
      notifyError,
      notifySuccess,
    ],
  );

  const handleHistoryEditClose = useCallback(() => {
    if (!isSubmitting) {
      if (onBack) {
        onBack();
      } else {
        setActiveTab("edit-history");
        setFormMode("create");
        setFormInitialData(null);
        setSelectedVersion(null);
      }
    }
  }, [isSubmitting, onBack]);

  const handleHistoryEditSubmit = useCallback(
    async (updateData: TemporalVersionPayload & { recordId: string }) => {
      setIsSubmitting(true);
      try {
        const lifecycleStatus =
          updateData.lifecycleStatus ?? "CURRENT";
        const mappedStatus = mapLifecycleStatusToApiStatus(lifecycleStatus);

        await updateHistoryRecord(
          organizationCode,
          updateData.recordId,
          {
            name: updateData.name,
            unitType: updateData.unitType,
            status: mappedStatus,
            lifecycleStatus,
            description: updateData.description ?? null,
            effectiveDate: updateData.effectiveDate,
            parentCode: normalizeParentCode.forAPI(updateData.parentCode),
            changeReason: "通过组织详情页面修改历史记录",
            operationReason: updateData.operationReason,
          },
        );

        await loadVersions(false, updateData.recordId);
        setActiveTab("edit-history");
        setFormMode("edit");
        setFormInitialData({
          name: updateData.name as string,
          unitType: updateData.unitType,
          status: mappedStatus,
          lifecycleStatus,
          description: (updateData.description as string) || "",
          parentCode: normalizeParentCode.forForm(
            updateData.parentCode as string,
          ),
          effectiveDate: updateData.effectiveDate as string,
        });
        setSelectedVersion((prev) => {
          if (!prev || prev.recordId !== updateData.recordId) {
            return prev;
          }
          return {
            ...prev,
            name: updateData.name as string,
            unitType: updateData.unitType as string,
            status: mappedStatus,
            lifecycleStatus,
            description: (updateData.description as string) || undefined,
            parentCode: (updateData.parentCode as string) || undefined,
            effectiveDate: updateData.effectiveDate as string,
          };
        });
        notifySuccess("历史记录修改成功！");
      } catch (historyError) {
        logger.error("修改历史记录失败:", historyError);
        notifyError("修改失败，请检查网络连接");
      } finally {
        setIsSubmitting(false);
      }
    },
    [organizationCode, loadVersions, notifyError, notifySuccess],
  );

  useEffect(() => {
    if (!isCreateMode && organizationCode) {
      loadVersions();
    }
  }, [loadVersions, isCreateMode, organizationCode]);

  useEffect(() => {
    if (organizationCode) {
      return;
    }

    setIsLoading(false);
    setFormMode("create");
    setFormInitialData(null);
    setSelectedVersion(null);
    setDisplayPaths(null);
  }, [organizationCode]);

  const currentOrganizationName = useMemo(() => {
    const currentVersion = versions.find((v) => v.isCurrent);
    return currentVersion?.name || "";
  }, [versions]);

  const currentTimelineStatus = useMemo(() => {
    const current = versions.find((v) => v.isCurrent);
    return current?.status || selectedVersion?.status;
  }, [versions, selectedVersion]);

  return [
    {
      versions,
      selectedVersion,
      isLoading,
      showDeleteConfirm,
      isDeleting,
      loadingError,
      successMessage,
      error,
      retryCount,
      isSubmitting,
      currentETag,
      activeTab,
      formMode,
      formInitialData,
      displayPaths,
      currentTimelineStatus,
      currentOrganizationName,
    },
    {
      setShowDeleteConfirm,
      loadVersions,
      handleStateMutationCompleted,
      handleDeleteVersion,
      handleVersionSelect,
      handleFormSubmit,
      handleHistoryEditClose,
      handleHistoryEditSubmit,
      setActiveTab,
      setCurrentETag,
      notifySuccess,
      notifyError,
    },
  ];
};
