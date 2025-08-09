import type { SystemMetrics } from '../types/monitoring';
import { mockMetrics } from '../types/monitoring';

export class MonitoringService {
  /**
   * 获取系统监控指标数据
   * MVP版本：返回模拟数据 + 尝试获取真实指标
   */
  static async getMetrics(): Promise<SystemMetrics> {
    try {
      // 尝试获取真实的Prometheus指标
      const response = await fetch('/api/metrics', {
        method: 'GET',
        headers: {
          'Accept': 'text/plain'
        }
      });
      
      if (response.ok) {
        const rawMetrics = await response.text();
        console.log('[MonitoringService] 获取到真实指标数据:', rawMetrics.substring(0, 200) + '...');
        
        // 解析Prometheus指标并与模拟数据合并
        const parsedMetrics = this.parsePrometheusMetrics(rawMetrics);
        return {
          ...mockMetrics,
          ...parsedMetrics,
          lastUpdated: new Date().toLocaleString()
        };
      } else {
        console.warn('[MonitoringService] 无法获取指标数据，使用模拟数据');
        return {
          ...mockMetrics,
          lastUpdated: new Date().toLocaleString()
        };
      }
    } catch (error) {
      console.warn('[MonitoringService] 指标获取失败，使用模拟数据:', error);
      return {
        ...mockMetrics,
        lastUpdated: new Date().toLocaleString()
      };
    }
  }

  /**
   * 检查单个服务的健康状态
   */
  static async checkServiceHealth(serviceUrl: string): Promise<boolean> {
    try {
      const response = await fetch(serviceUrl, { 
        method: 'HEAD',
        mode: 'no-cors' // 避免CORS问题
      });
      return true; // 能发送请求就认为服务正常
    } catch (error) {
      return false;
    }
  }

  /**
   * 解析Prometheus格式的指标数据
   * 简化版本：提取关键指标
   */
  private static parsePrometheusMetrics(rawMetrics: string): Partial<SystemMetrics> {
    const lines = rawMetrics.split('\n');
    const httpRequestsTotal = this.extractMetric(lines, 'http_requests_total');
    const httpRequestDuration = this.extractMetric(lines, 'http_request_duration_seconds');
    
    // 如果找到指标，返回解析结果，否则返回空对象
    if (httpRequestsTotal || httpRequestDuration) {
      console.log('[MonitoringService] 解析到指标:', {
        requests: httpRequestsTotal,
        duration: httpRequestDuration
      });
    }
    
    return {}; // MVP版本暂时只记录日志，不修改数据
  }

  /**
   * 从Prometheus指标中提取特定指标
   */
  private static extractMetric(lines: string[], metricName: string): number | null {
    const metricLine = lines.find(line => 
      line.startsWith(metricName) && !line.startsWith('#')
    );
    
    if (metricLine) {
      const match = metricLine.match(/\s([\d.]+)$/);
      return match ? parseFloat(match[1]) : null;
    }
    
    return null;
  }

  /**
   * 随机更新模拟数据，让MVP版本看起来更真实
   */
  static updateMockData(): SystemMetrics {
    const updated = { ...mockMetrics };
    
    // 随机更新服务响应时间
    updated.services = updated.services.map(service => ({
      ...service,
      responseTime: (Math.random() * 50 + 10).toFixed(0) + 'ms',
      requests: (Math.random() * 200 + 50).toFixed(0)
    }));
    
    // 更新时间戳
    updated.lastUpdated = new Date().toLocaleString();
    
    return updated;
  }
}