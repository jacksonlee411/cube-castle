/**
 * 时间日期选择器组件
 * 用于选择历史查看时点和时间范围
 */
import React, { useState, useCallback, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Modal } from '@workday/canvas-kit-react/modal';
import { Card } from '@workday/canvas-kit-react/card';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { colors, space, borderRadius } from '@workday/canvas-kit-react/tokens';

export interface DateTimePickerProps {
  /** 是否显示弹窗 */
  isOpen: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 日期选择回调 */
  onSelect: (dateTime: string) => void;
  /** 默认日期 */
  defaultDate?: string;
  /** 标题 */
  title?: string;
  /** 是否显示时间选择 */
  showTime?: boolean;
  /** 是否显示预设选项 */
  showPresets?: boolean;
  /** 最小日期 */
  minDate?: string;
  /** 最大日期 */
  maxDate?: string;
}

/**
 * 时间日期选择器组件
 */
export const DateTimePicker: React.FC<DateTimePickerProps> = ({
  isOpen,
  onClose,
  onSelect,
  defaultDate,
  title = '选择日期时间',
  showTime = true,
  showPresets = true,
  minDate,
  maxDate
}) => {
  const [selectedDate, setSelectedDate] = useState('');
  const [selectedTime, setSelectedTime] = useState('');
  const [customInput, setCustomInput] = useState('');

  // 初始化日期时间
  useEffect(() => {
    if (defaultDate) {
      try {
        const date = new Date(defaultDate);
        setSelectedDate(date.toISOString().split('T')[0]);
        setSelectedTime(date.toTimeString().slice(0, 5));
        setCustomInput(defaultDate);
      } catch {
        // 如果默认日期无效，使用当前时间
        const now = new Date();
        setSelectedDate(now.toISOString().split('T')[0]);
        setSelectedTime(now.toTimeString().slice(0, 5));
      }
    } else {
      const now = new Date();
      setSelectedDate(now.toISOString().split('T')[0]);
      setSelectedTime(now.toTimeString().slice(0, 5));
    }
  }, [defaultDate]);

  // 预设选项
  const presetOptions = [
    {
      label: '现在',
      value: () => new Date().toISOString(),
      description: '当前时间'
    },
    {
      label: '今天开始',
      value: () => {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        return today.toISOString();
      },
      description: '今天 00:00'
    },
    {
      label: '1小时前',
      value: () => new Date(Date.now() - 60 * 60 * 1000).toISOString(),
      description: '1小时前'
    },
    {
      label: '1天前',
      value: () => new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      description: '昨天此时'
    },
    {
      label: '1周前',
      value: () => new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
      description: '上周此时'
    },
    {
      label: '1个月前',
      value: () => new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
      description: '上个月此时'
    },
    {
      label: '3个月前',
      value: () => new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString(),
      description: '3个月前此时'
    },
    {
      label: '1年前',
      value: () => new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString(),
      description: '去年此时'
    }
  ];

  // 处理预设选项点击
  const handlePresetClick = useCallback((preset: typeof presetOptions[0]) => {
    const dateTime = preset.value();
    const date = new Date(dateTime);
    setSelectedDate(date.toISOString().split('T')[0]);
    setSelectedTime(date.toTimeString().slice(0, 5));
    setCustomInput(dateTime);
  }, []);

  // 处理日期变更
  const handleDateChange = useCallback((date: string) => {
    setSelectedDate(date);
    updateCustomInput(date, selectedTime);
  }, [selectedTime]);

  // 处理时间变更
  const handleTimeChange = useCallback((time: string) => {
    setSelectedTime(time);
    updateCustomInput(selectedDate, time);
  }, [selectedDate]);

  // 更新自定义输入
  const updateCustomInput = useCallback((date: string, time: string) => {
    if (date) {
      const dateTime = showTime && time ? `${date}T${time}:00.000Z` : `${date}T00:00:00.000Z`;
      setCustomInput(dateTime);
    }
  }, [showTime]);

  // 处理自定义输入变更
  const handleCustomInputChange = useCallback((value: string) => {
    setCustomInput(value);
    
    // 尝试解析日期时间
    try {
      const date = new Date(value);
      if (!isNaN(date.getTime())) {
        setSelectedDate(date.toISOString().split('T')[0]);
        setSelectedTime(date.toTimeString().slice(0, 5));
      }
    } catch {
      // 忽略解析错误
    }
  }, []);

  // 处理确认选择
  const handleConfirm = useCallback(() => {
    try {
      let dateTime: string;
      
      if (customInput) {
        // 验证自定义输入
        const date = new Date(customInput);
        if (isNaN(date.getTime())) {
          throw new Error('无效的日期时间格式');
        }
        dateTime = date.toISOString();
      } else {
        // 从选择的日期时间构建
        const dateTimeStr = showTime && selectedTime 
          ? `${selectedDate}T${selectedTime}:00.000Z`
          : `${selectedDate}T00:00:00.000Z`;
        dateTime = new Date(dateTimeStr).toISOString();
      }

      // 验证日期范围
      if (minDate && new Date(dateTime) < new Date(minDate)) {
        throw new Error('选择的日期时间早于最小允许日期');
      }
      
      if (maxDate && new Date(dateTime) > new Date(maxDate)) {
        throw new Error('选择的日期时间晚于最大允许日期');
      }

      onSelect(dateTime);
    } catch (error) {
      alert(error instanceof Error ? error.message : '日期时间选择错误');
    }
  }, [customInput, selectedDate, selectedTime, showTime, minDate, maxDate, onSelect]);

  // 格式化显示时间
  const formatPreviewTime = (dateTimeStr: string) => {
    try {
      const date = new Date(dateTimeStr);
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      });
    } catch {
      return dateTimeStr;
    }
  };

  if (!isOpen) {
    return null;
  }

  return (
    <Modal onClose={onClose}>
      <Card 
        padding={space.l}
        minWidth="500px"
        maxWidth="600px"
      >
        <Box marginBottom={space.m}>
          <Text fontSize="large" fontWeight="bold">
            {title}
          </Text>
        </Box>

        <Flex gap={space.l}>
          {/* 预设选项 */}
          {showPresets && (
            <Box flex="1">
              <Text fontSize="medium" fontWeight="medium" marginBottom={space.s}>
                快速选择
              </Text>
              <Box
                backgroundColor={colors.soap100}
                borderRadius={borderRadius.m}
                padding={space.s}
                maxHeight="300px"
                overflow="auto"
              >
                {presetOptions.map((preset, index) => (
                  <SecondaryButton
                    key={index}
                    variant="plain"
                    size="small"
                    onClick={() => handlePresetClick(preset)}
                    style={{
                      display: 'block',
                      width: '100%',
                      textAlign: 'left',
                      marginBottom: space.xxxs,
                      padding: space.xs
                    }}
                  >
                    <Flex justifyContent="space-between" alignItems="center">
                      <Text fontSize="small" fontWeight="medium">
                        {preset.label}
                      </Text>
                      <Text fontSize="small" color={colors.licorice500}>
                        {preset.description}
                      </Text>
                    </Flex>
                  </SecondaryButton>
                ))}
              </Box>
            </Box>
          )}

          <Divider orientation="vertical" />

          {/* 自定义选择 */}
          <Box flex="1">
            <Text fontSize="medium" fontWeight="medium" marginBottom={space.s}>
              自定义选择
            </Text>
            
            <Box marginBottom={space.m}>
              <Text fontSize="small" marginBottom={space.xs}>
                📅 日期
              </Text>
              <TextInput
                type="date"
                value={selectedDate}
                onChange={(e) => handleDateChange(e.target.value)}
                min={minDate?.split('T')[0]}
                max={maxDate?.split('T')[0]}
              />
            </Box>

            {showTime && (
              <Box marginBottom={space.m}>
                <Text fontSize="small" marginBottom={space.xs}>
                  🕰️ 时间
                </Text>
                <TextInput
                  type="time"
                  value={selectedTime}
                  onChange={(e) => handleTimeChange(e.target.value)}
                />
              </Box>
            )}

            <Box marginBottom={space.m}>
              <Text fontSize="small" marginBottom={space.xs}>
                ISO 8601 格式 (高级)
              </Text>
              <TextInput
                value={customInput}
                onChange={(e) => handleCustomInputChange(e.target.value)}
                placeholder="2024-01-01T12:00:00.000Z"
              />
            </Box>

            {/* 预览 */}
            {customInput && (
              <Box
                backgroundColor={colors.soap100}
                borderRadius={borderRadius.s}
                padding={space.s}
                marginBottom={space.m}
              >
                <Text fontSize="small" color={colors.licorice600}>
                  预览: {formatPreviewTime(customInput)}
                </Text>
              </Box>
            )}
          </Box>
        </Flex>

        {/* 操作按钮 */}
        <Flex justifyContent="flex-end" gap={space.s} marginTop={space.l}>
          <SecondaryButton variant="secondary" onClick={onClose}>
            取消
          </SecondaryButton>
          <PrimaryButton onClick={handleConfirm}>
            确认选择
          </PrimaryButton>
        </Flex>
      </Card>
    </Modal>
  );
};

export default DateTimePicker;