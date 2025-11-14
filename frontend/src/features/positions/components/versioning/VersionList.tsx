import React from 'react';
import { Card } from '@workday/canvas-kit-react/card';
import { Table } from '@workday/canvas-kit-react/table';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import type { PositionRecord } from '@/shared/types/positions';
import { getVersionKey } from './utils';
import { getPositionStatusMeta } from '@/features/temporal/entity/statusMeta';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

export interface PositionVersionListProps {
  versions: PositionRecord[];
  isLoading?: boolean;
  selectedVersionKey?: string | null;
  onSelectVersion?: (version: PositionRecord, versionKey: string) => void;
  versionKeys?: string[];
}

const formatVersionLabel = (version: PositionRecord): string => {
  const statusMeta = getPositionStatusMeta(version.status);
  const badges: string[] = [];

  if (version.isCurrent) {
    badges.push('当前');
  }
  if (version.isFuture) {
    badges.push('计划');
  }
  if (version.status === 'DELETED') {
    badges.push('已删除');
  }

  const badgeText = badges.length > 0 ? ` · ${badges.join(' / ')}` : '';
  return `${version.effectiveDate} · ${statusMeta.label}${badgeText}`;
};

export const PositionVersionList: React.FC<PositionVersionListProps> = ({
  versions,
  isLoading = false,
  selectedVersionKey,
  onSelectVersion,
  versionKeys,
}) => {
  return (
    <Card
      padding={space.l}
      backgroundColor={colors.frenchVanilla100}
      data-testid={temporalEntitySelectors.position.versionList}
    >
      <Heading size="small" marginBottom={space.m}>
        职位版本记录
      </Heading>

      {isLoading ? (
        <Text color={colors.licorice400}>正在加载职位版本...</Text>
      ) : versions.length === 0 ? (
        <Text color={colors.licorice400}>暂无职位版本记录</Text>
      ) : (
        <Table>
          <Table.Head>
            <Table.Row>
              <Table.Header width="220px">版本信息</Table.Header>
              <Table.Header width="160px">生效日期</Table.Header>
              <Table.Header width="140px">结束日期</Table.Header>
              <Table.Header width="120px">状态</Table.Header>
              <Table.Header width="200px">更新时间</Table.Header>
            </Table.Row>
          </Table.Head>
          <Table.Body>
            {versions.map((version, index) => {
              const key = versionKeys?.[index] ?? getVersionKey(version);
              const isSelected = selectedVersionKey === key;

              return (
                <Table.Row
                  key={key}
                  data-testid={`position-version-row-${key}`}
                  onClick={() => {
                    if (onSelectVersion) {
                      onSelectVersion(version, key);
                    }
                  }}
                  style={{
                    cursor: onSelectVersion ? 'pointer' : 'default',
                    backgroundColor: isSelected ? colors.soap200 : 'inherit',
                  }}
                >
                  <Table.Cell>
                    <Text fontWeight={isSelected ? 'bold' : 'normal'}>{formatVersionLabel(version)}</Text>
                  </Table.Cell>
                  <Table.Cell>{version.effectiveDate}</Table.Cell>
                  <Table.Cell>{version.endDate ?? '—'}</Table.Cell>
                  <Table.Cell>
                    <StatusPill status={version.status} />
                  </Table.Cell>
                  <Table.Cell>{version.updatedAt}</Table.Cell>
                </Table.Row>
              );
            })}
          </Table.Body>
        </Table>
      )}
    </Card>
  );
};

export default PositionVersionList;
const StatusPill: React.FC<{ status: string }> = ({ status }) => {
  const meta = getPositionStatusMeta(status);

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '4px 8px',
        borderRadius: 12,
        fontSize: 12,
        fontWeight: 600,
        color: meta.color,
        backgroundColor: meta.background,
        border: `1px solid ${meta.border}`,
      }}
    >
      {meta.label}
    </span>
  );
};
