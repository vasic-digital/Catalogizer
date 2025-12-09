import React from 'react';
import { cn } from '@/lib/utils';

interface TextareaProps {
  label?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  placeholder?: string;
  rows?: number;
  className?: string;
  disabled?: boolean;
  error?: string;
}

export const Textarea: React.FC<TextareaProps> = ({
  label,
  value = '',
  onChange,
  placeholder,
  rows = 4,
  className = '',
  disabled = false,
  error,
}) => {
  return (
    <div className={cn('mb-4', className)}>
      {label && (
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          {label}
        </label>
      )}
      <textarea
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        rows={rows}
        disabled={disabled}
        className={cn(
          'w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg',
          'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent',
          'bg-white dark:bg-gray-800 text-gray-900 dark:text-white',
          'resize-none',
          disabled && 'opacity-50 cursor-not-allowed',
          error && 'border-red-500 focus:ring-red-500',
          className
        )}
      />
      {error && (
        <p className="mt-1 text-sm text-red-600 dark:text-red-400">{error}</p>
      )}
    </div>
  );
};