import React from "react";
import { Box, Flex } from "@workday/canvas-kit-react/layout";
import { Heading, Text } from "@workday/canvas-kit-react/text";
import { SecondaryButton } from "@workday/canvas-kit-react/button";
import { useNavigate } from "react-router-dom";
import { OrganizationBreadcrumb } from "../../../shared/components/OrganizationBreadcrumb";
import { SuspendActivateButtons } from "./SuspendActivateButtons";
import type { OrganizationStateMutationResult } from "../../../shared/hooks/useOrganizationMutations";

interface TemporalMasterDetailHeaderProps {
  isCreateMode: boolean;
  organizationCode: string | null;
  organizationName: string;
  displayPaths: { codePath: string; namePath: string } | null;
  isLoading: boolean;
  isSubmitting: boolean;
  readonly?: boolean;
  currentTimelineStatus: string | undefined;
  currentETag: string | null;
  onRefresh: () => void;
  onETagChange: (etag: string | null) => void;
  onSuccess: (message: string) => void;
  onError: (message: string) => void;
  onCompleted: (
    action: "suspend" | "activate",
    result: OrganizationStateMutationResult,
  ) => Promise<void>;
}

export const TemporalMasterDetailHeader: React.FC<TemporalMasterDetailHeaderProps> = ({
  isCreateMode,
  organizationCode,
  organizationName,
  displayPaths,
  isLoading,
  isSubmitting,
  readonly = false,
  currentTimelineStatus,
  currentETag,
  onRefresh,
  onETagChange,
  onSuccess,
  onError,
  onCompleted,
}) => {
  const navigate = useNavigate();

  const title = isCreateMode
    ? "新建组织 - 编辑组织信息"
    : `组织详情 - ${organizationCode}${organizationName ? ` ${organizationName}` : ""}`;

  const subtitle = isCreateMode
    ? "填写组织基本信息，系统将自动分配组织编码"
    : "强制时间连续性的组织架构管理";

  return (
    <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
      <Box>
        <Heading size="large">{title}</Heading>
        <Text typeLevel="subtext.medium" color="hint">
          {subtitle}
        </Text>
        {!isCreateMode && displayPaths && (
          <Box marginTop="s">
            <OrganizationBreadcrumb
              codePath={displayPaths.codePath}
              namePath={displayPaths.namePath}
              separator="/"
              onNavigate={(code) => {
                if (code) {
                  navigate(`/organizations/${code}/temporal`);
                }
              }}
            />
          </Box>
        )}
      </Box>

      <Flex gap="s" alignItems="center">
        {!isCreateMode && organizationCode && (
          <SuspendActivateButtons
            organizationCode={organizationCode}
            currentStatus={currentTimelineStatus}
            currentETag={currentETag}
            readonly={readonly}
            disabled={isLoading || isSubmitting}
            onETagChange={onETagChange}
            onSuccess={onSuccess}
            onError={onError}
            onCompleted={onCompleted}
          />
        )}
        <SecondaryButton onClick={onRefresh} disabled={isLoading}>
          {isLoading ? "刷新中..." : "刷新"}
        </SecondaryButton>
      </Flex>
    </Flex>
  );
};

export default TemporalMasterDetailHeader;
