import React from 'react';

export type SummaryColor = 'blue' | 'green' | 'orange' | 'purple' | 'indigo';

interface SummaryStatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: React.ReactNode;
  color?: SummaryColor;
}

const colorToGradient: Record<SummaryColor, string> = {
  blue: 'from-blue-500 to-blue-600',
  green: 'from-green-500 to-emerald-600',
  orange: 'from-orange-500 to-amber-600',
  purple: 'from-purple-500 to-violet-600',
  indigo: 'from-indigo-500 to-blue-600',
};

export const SummaryStatCard: React.FC<SummaryStatCardProps> = ({
  title,
  value,
  subtitle,
  icon,
  color = 'blue',
}) => {
  const gradient = colorToGradient[color] || colorToGradient.blue;

  return (
    <div className={`rounded-xl p-5 text-white shadow-md bg-gradient-to-r ${gradient}`}>
      <div className="flex items-center">
        {icon && (
          <div className="mr-4 text-white/90">
            {icon}
          </div>
        )}
        <div>
          <div className="text-sm font-medium opacity-90">{title}</div>
          <div className="text-2xl font-bold leading-tight">{value}</div>
          {subtitle && (
            <div className="text-xs opacity-90 mt-1">{subtitle}</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SummaryStatCard;