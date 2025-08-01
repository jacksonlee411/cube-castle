// Production-safe logging utility with SWR monitoring
type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface Logger {
  debug: (message: string, ...args: any[]) => void;
  info: (message: string, ...args: any[]) => void;
  warn: (message: string, ...args: any[]) => void;
  error: (message: string, ...args: any[]) => void;
}

interface MetricsData {
  component: string;
  action: string;
  success: boolean;
  duration: number;
  error?: string;
  timestamp: number;
}

class ProductionLogger implements Logger {
  private isDevelopment = process.env.NODE_ENV === 'development';
  private metrics: MetricsData[] = [];
  private maxMetricsSize = 100;

  debug(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.debug(message, ...args);
    }
  }

  info(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.info(message, ...args);
    }
    // In production, you might want to send to logging service
  }

  warn(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.warn(message, ...args);
    } else {
      // In production, send to error tracking service like Sentry
      // Example: Sentry.captureMessage(message, 'warning');
    }
  }

  error(message: string, ...args: any[]): void {
    if (this.isDevelopment) {
      console.error(message, ...args);
    } else {
      // In production, send to error tracking service
      // Example: Sentry.captureException(new Error(message));
    }
  }

  // SWR monitoring methods
  private addMetric(metric: MetricsData): void {
    this.metrics.push(metric);
    
    // Keep metrics array size limited
    if (this.metrics.length > this.maxMetricsSize) {
      this.metrics = this.metrics.slice(-this.maxMetricsSize);
    }
  }

  trackSWRRequest(key: string, success: boolean, duration: number, error?: Error): void {
    this.addMetric({
      component: 'SWR',
      action: key,
      success,
      duration,
      error: error?.message,
      timestamp: Date.now()
    });

    if (this.isDevelopment) {
      const status = success ? '✅' : '❌';
      this.info(`SWR ${status} ${key} (${duration}ms)`, { success, duration, error: error?.message });
    }
  }

  getMetrics(component?: string, action?: string): MetricsData[] {
    let filtered = this.metrics;
    
    if (component) {
      filtered = filtered.filter(m => m.component === component);
    }
    
    if (action) {
      filtered = filtered.filter(m => m.action === action);
    }
    
    return filtered;
  }

  getSuccessRate(component?: string, action?: string): { rate: number; total: number; successful: number } {
    const metrics = this.getMetrics(component, action);
    const total = metrics.length;
    const successful = metrics.filter(m => m.success).length;
    
    return {
      rate: total > 0 ? successful / total : 0,
      total,
      successful
    };
  }

  getAverageResponseTime(component?: string, action?: string): number {
    const metrics = this.getMetrics(component, action)
      .filter(m => m.success); // Only successful requests
      
    if (metrics.length === 0) return 0;
    
    const totalDuration = metrics.reduce((sum, m) => sum + m.duration, 0);
    return totalDuration / metrics.length;
  }

  getSWRMetrics(key?: string) {
    const component = 'SWR';
    const action = key;
    
    return {
      successRate: this.getSuccessRate(component, action),
      averageResponseTime: this.getAverageResponseTime(component, action),
      recentMetrics: this.getMetrics(component, action).slice(-20)
    };
  }

  getPerformanceInsights(): {
    slowestEndpoints: Array<{ key: string; avgTime: number }>;
    errorProneEndpoints: Array<{ key: string; errorRate: number }>;
    overallHealth: { successRate: number; avgResponseTime: number };
  } {
    const swrMetrics = this.getMetrics('SWR');
    const endpointGroups = swrMetrics.reduce((acc, metric) => {
      if (!acc[metric.action]) {
        acc[metric.action] = [];
      }
      acc[metric.action].push(metric);
      return acc;
    }, {} as Record<string, MetricsData[]>);

    const slowestEndpoints = Object.entries(endpointGroups)
      .map(([key, metrics]) => ({
        key,
        avgTime: metrics.filter(m => m.success).reduce((sum, m) => sum + m.duration, 0) / metrics.filter(m => m.success).length || 0
      }))
      .sort((a, b) => b.avgTime - a.avgTime)
      .slice(0, 5);

    const errorProneEndpoints = Object.entries(endpointGroups)
      .map(([key, metrics]) => ({
        key,
        errorRate: metrics.filter(m => !m.success).length / metrics.length
      }))
      .filter(endpoint => endpoint.errorRate > 0)
      .sort((a, b) => b.errorRate - a.errorRate)
      .slice(0, 5);

    const overallStats = this.getSuccessRate('SWR');
    const overallResponseTime = this.getAverageResponseTime('SWR');

    return {
      slowestEndpoints,
      errorProneEndpoints,
      overallHealth: {
        successRate: overallStats.rate,
        avgResponseTime: overallResponseTime
      }
    };
  }

  clearMetrics(): void {
    this.metrics = [];
  }
}

export const logger = new ProductionLogger();

// For backward compatibility and migration
export const log = {
  debug: logger.debug.bind(logger),
  info: logger.info.bind(logger),
  warn: logger.warn.bind(logger),
  error: logger.error.bind(logger),
};