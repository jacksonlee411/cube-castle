export interface StatCardProps {
  title: string;
  stats: Record<string, number>;
  variant?: 'default' | 'highlight';
}

export interface StatsCardsProps {
  stats: {
    byType: Record<string, number>;
    byStatus: Record<string, number>;
    totalCount: number;
  };
}