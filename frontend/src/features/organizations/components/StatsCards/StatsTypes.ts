export interface StatCardProps {
  title: string;
  stats: Record<string, number>;
  variant?: 'default' | 'highlight';
}

export interface StatsCardsProps {
  stats: {
    by_type: Record<string, number>;
    by_status: Record<string, number>;
    total_count: number;
  };
}