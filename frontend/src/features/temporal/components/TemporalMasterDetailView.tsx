/**
 * 组织详情主从视图组件
 * 左侧：垂直交互式时间轴导航
 * 右侧：动态版本详情卡片
 * 实现强制时间连续性的完整用户体验
 */
import React from "react";
import { Box, Flex } from "@workday/canvas-kit-react/layout";
import { Text } from "@workday/canvas-kit-react/text";
import {
  PrimaryButton,
  SecondaryButton,
} from "@workday/canvas-kit-react/button";
import { Card } from "@workday/canvas-kit-react/card";
import { Modal, useModalModel } from "@workday/canvas-kit-react/modal";
import { activityStreamIcon } from "@workday/canvas-system-icons-web";
import { SystemIcon } from "@workday/canvas-kit-react/icon";
import { InlineNewVersionForm } from "./InlineNewVersionForm";
import { TimelineComponent } from "./TimelineComponent";
import { TabNavigation } from "./TabNavigation";
import { colors, borderRadius } from "@workday/canvas-kit-react/tokens";
import { baseColors } from "../../../shared/utils/colorTokens";
// 审计历史组件导入
import { AuditHistorySection } from "../../audit/components/AuditHistorySection";
import {
  useTemporalMasterDetail,
  type TemporalMasterDetailViewProps,
} from "./hooks/useTemporalMasterDetail";
import TemporalMasterDetailHeader from "./TemporalMasterDetailHeader";
import TemporalMasterDetailAlerts from "./TemporalMasterDetailAlerts";

export type { TemporalMasterDetailViewProps } from "./hooks/useTemporalMasterDetail";

/**
 * 组织详情主从视图主组件
 */
export const TemporalMasterDetailView: React.FC<
  TemporalMasterDetailViewProps
> = ({
  organizationCode,
  readonly = false,
  onBack,
  onCreateSuccess,
  isCreateMode = false,
}) => {
  const [state, handlers] = useTemporalMasterDetail({
    organizationCode,
    readonly,
    onBack,
    onCreateSuccess,
    isCreateMode,
  });

  const {
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
    earliestVersion,
    isEarliestVersionSelected,
  } = state;

  const {
    setShowDeleteConfirm,
    loadVersions,
    handleStateMutationCompleted,
    handleDeleteOrganization,
    handleDeleteVersion,
    handleVersionSelect,
    handleFormSubmit,
    handleHistoryEditClose,
    handleHistoryEditSubmit,
    setActiveTab: updateActiveTab,
    setCurrentETag,
    notifySuccess,
    notifyError,
  } = handlers;

  const deleteModalModel = useModalModel();

  React.useEffect(() => {
    if (showDeleteConfirm && deleteModalModel.state.visibility !== "visible") {
      deleteModalModel.events.show();
    } else if (
      !showDeleteConfirm &&
      deleteModalModel.state.visibility === "visible"
    ) {
      deleteModalModel.events.hide();
    }
  }, [showDeleteConfirm, deleteModalModel]);

  return (
    <Box padding="l" data-testid="temporal-master-detail-view">
      <TemporalMasterDetailHeader
        isCreateMode={isCreateMode}
        organizationCode={organizationCode}
        organizationName={currentOrganizationName}
        displayPaths={displayPaths}
        isLoading={isLoading}
        isSubmitting={isSubmitting}
        readonly={readonly}
        currentTimelineStatus={currentTimelineStatus}
        currentETag={currentETag}
        onRefresh={() => loadVersions()}
        onETagChange={setCurrentETag}
        onSuccess={notifySuccess}
        onError={notifyError}
        onCompleted={handleStateMutationCompleted}
      />

      <TemporalMasterDetailAlerts
        loadingError={loadingError}
        error={error}
        successMessage={successMessage}
        retryCount={retryCount}
        isLoading={isLoading}
        onRetry={() => loadVersions(true)}
      />

      {/* 主从视图布局 */}
      <Flex gap="l" height="calc(100vh - 220px)">
        {/* 左侧：垂直交互式时间轴导航 */}
        {!isCreateMode && (
          <TimelineComponent
            versions={versions}
            selectedVersion={selectedVersion}
            onVersionSelect={handleVersionSelect}
            onDeleteVersion={
              readonly ? undefined : (version) => setShowDeleteConfirm(version)
            }
            isLoading={isLoading}
            readonly={readonly}
          />
        )}

        {/* 创建模式下的提示区域 */}
        {isCreateMode && (
          <Box
            width="350px"
            height="calc(100vh - 200px)"
            backgroundColor="#F8F9FA"
            borderRadius={borderRadius.m}
            border="1px solid #E9ECEF"
            padding="m"
            style={{
              display: "flex",
              flexDirection: "column",
              justifyContent: "center",
              alignItems: "center",
            }}
          >
            <Box textAlign="center">
              <Text typeLevel="heading.small" marginBottom="m">
                创建新组织
              </Text>
              <Text typeLevel="body.medium" color="hint" marginBottom="l">
                填写右侧表单信息后，系统将自动分配组织编码并生成首个时态记录
              </Text>
              <Box
                width="60px"
                height="60px"
                borderRadius="50%"
                backgroundColor={colors.blueberry600}
                margin="auto"
                style={{
                  display: "flex",
                  justifyContent: "center",
                  alignItems: "center",
                }}
              >
                <Text color="white" typeLevel="heading.medium">
                  +
                </Text>
              </Box>
            </Box>
          </Box>
        )}

        {/* 右侧：选项卡视图 */}
        <Box flex="1">
          {isCreateMode ? (
            // 创建模式：直接显示创建表单
            <InlineNewVersionForm
              organizationCode={null} // 创建模式下传入null
              onSubmit={handleFormSubmit}
              onCancel={() => {
                if (onBack) {
                  onBack(); // 创建模式下取消应该返回上一页
                }
              }}
              isSubmitting={isSubmitting}
              mode={formMode}
              initialData={formMode === "edit" ? formInitialData : null}
              selectedVersion={formMode === "edit" ? selectedVersion : null}
              allVersions={null} // 创建模式不需要版本数据
              hierarchyPaths={displayPaths}
            />
          ) : (
            // 正常模式：带选项卡的多功能视图
            <>
              {/* 选项卡导航 */}
              <TabNavigation
                activeTab={activeTab}
                onTabChange={updateActiveTab}
                disabled={isSubmitting || isLoading}
                tabs={[
                  { key: "edit-history", label: "版本历史" },
                  { key: "audit-history", label: "审计历史" },
                ]}
              />

              {/* 选项卡内容 */}
              {activeTab === "edit-history" && (
                <InlineNewVersionForm
                  organizationCode={organizationCode}
                  onSubmit={handleFormSubmit}
                  onCancel={handleHistoryEditClose}
                  isSubmitting={isSubmitting}
                  mode={formMode}
                  initialData={formMode === "edit" ? formInitialData : null}
                  selectedVersion={formMode === "edit" ? selectedVersion : null}
                  allVersions={versions.map((v) => ({
                    // 传递版本数据用于日期范围验证
                    recordId: v.recordId,
                    effectiveDate: v.effectiveDate,
                    endDate: v.endDate,
                    isCurrent: v.isCurrent,
                  }))}
                  onEditHistory={handleHistoryEditSubmit}
                  onDeactivate={async (version) => {
                    try {
                      await handleDeleteVersion(version);
                    } catch (error) {
                      const errorMessage =
                        error instanceof Error
                          ? error.message
                          : "作废失败，请重试";
                      notifyError(errorMessage);
                    }
                  }} // 传递作废功能
                  onInsertRecord={handleFormSubmit} // 传递插入记录功能
                  activeTab="edit-history"
                  onTabChange={updateActiveTab}
                  hierarchyPaths={displayPaths}
                  canDeleteOrganization={!readonly && isEarliestVersionSelected}
                  onDeleteOrganization={
                    earliestVersion
                      ? () => handleDeleteOrganization(earliestVersion)
                      : undefined
                  }
                  isDeletingOrganization={isDeleting}
                />
              )}

              {/* 审计历史标签页 */}
              {activeTab === "audit-history" && selectedVersion?.recordId && (
                <>
                  {/* 调试信息 */}
                  <Box
                    marginBottom="s"
                    padding="s"
                    backgroundColor="#f5f5f5"
                    borderRadius="4px"
                  >
                    <Text typeLevel="subtext.small" color="hint">
                      🔍 调试信息: recordId = {selectedVersion.recordId}
                    </Text>
                  </Box>
                  <AuditHistorySection
                    recordId={selectedVersion.recordId}
                    params={{
                      limit: 50,
                      mode: "current",
                    }}
                  />
                </>
              )}

              {activeTab === "audit-history" && !selectedVersion?.recordId && (
                <Card padding="m">
                  <Flex alignItems="center" gap="xs" marginBottom="m">
                    <SystemIcon icon={activityStreamIcon} size={16} />
                    <Text as="h3" typeLevel="subtext.large" fontWeight="bold">
                      审计历史
                    </Text>
                  </Flex>
                  <Text typeLevel="body.medium" color="hint">
                    请选择一个版本查看对应的审计历史记录
                  </Text>
                </Card>
              )}
            </>
          )}
        </Box>
      </Flex>

      {/* 作废确认对话框 */}
      {showDeleteConfirm && (
        <Modal model={deleteModalModel}>
          <Modal.Overlay>
            <Modal.Card>
              <Modal.CloseIcon onClick={() => setShowDeleteConfirm(null)} />
              <Modal.Heading>确认作废版本</Modal.Heading>
              <Modal.Body>
                <Box padding="l">
                  <Flex alignItems="flex-start" gap="m" marginBottom="l">
                    <Box fontSize="24px" color={baseColors.cinnamon[600]}>
                      警告
                    </Box>
                    <Box>
                      <Text typeLevel="body.medium" marginBottom="s">
                        确定要作废生效日期为{" "}
                        <strong>
                          {new Date(
                            showDeleteConfirm.effectiveDate,
                          ).toLocaleDateString("zh-CN")}
                        </strong>{" "}
                        的版本吗？
                      </Text>
                      <Text
                        typeLevel="subtext.small"
                        color="hint"
                        marginBottom="s"
                      >
                        版本名称: {showDeleteConfirm.name}
                      </Text>
                      <Text
                        typeLevel="subtext.small"
                        color={baseColors.cinnamon[600]}
                      >
                        警告 作废后将自动填补时间空洞，此操作不可撤销
                      </Text>
                    </Box>
                  </Flex>

                  <Flex gap="s" justifyContent="flex-end">
                    <SecondaryButton
                      onClick={() => setShowDeleteConfirm(null)}
                      disabled={isDeleting}
                    >
                      取消
                    </SecondaryButton>
                    <PrimaryButton
                      onClick={async () => {
                        try {
                          await handleDeleteVersion(showDeleteConfirm);
                          // 成功时的处理由handleDeleteVersion内部完成
                        } catch (error) {
                          // 显示错误消息
                          const errorMessage =
                            error instanceof Error
                              ? error.message
                              : "作废失败，请重试";
                          notifyError(errorMessage);
                        }
                      }}
                      disabled={isDeleting}
                    >
                      {isDeleting ? "作废中..." : "确认作废"}
                    </PrimaryButton>
                  </Flex>
                </Box>
              </Modal.Body>
            </Modal.Card>
          </Modal.Overlay>
        </Modal>
      )}
    </Box>
  );
};

export default TemporalMasterDetailView;
