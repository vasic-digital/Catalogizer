import React, { useState, useEffect } from 'react';
import { AdminPanel } from '@/components/admin/AdminPanel';
import { adminApi } from '@/lib/adminApi';
import { useQuery } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { SystemInfo, User, StorageInfo, Backup } from '@/types/admin';

export const Admin: React.FC = () => {
  const { data: systemInfo, isLoading: systemLoading } = useQuery({
    queryKey: ['admin-system-info'],
    queryFn: () => adminApi.getSystemInfo(),
    staleTime: 1000 * 60 * 2,
    refetchInterval: 1000 * 30, // Refresh every 30 seconds
  });

  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['admin-users'],
    queryFn: () => adminApi.getUsers(),
    staleTime: 1000 * 60 * 5,
  });

  const { data: storageInfo, isLoading: storageLoading } = useQuery({
    queryKey: ['admin-storage'],
    queryFn: () => adminApi.getStorageInfo(),
    staleTime: 1000 * 60 * 5,
  });

  const { data: backups, isLoading: backupsLoading } = useQuery({
    queryKey: ['admin-backups'],
    queryFn: () => adminApi.getBackups(),
    staleTime: 1000 * 60 * 2,
  });

  const handleCreateBackup = async (type: 'full' | 'incremental') => {
    try {
      await adminApi.createBackup(type);
      toast.success(`${type === 'full' ? 'Full' : 'Incremental'} backup started successfully`);
    } catch (error) {
      toast.error(`Failed to create backup: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleRestoreBackup = async (id: string) => {
    try {
      await adminApi.restoreBackup(id);
      toast.success('Backup restore initiated successfully');
    } catch (error) {
      toast.error(`Failed to restore backup: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleScanStorage = async (path: string) => {
    try {
      await adminApi.scanStorage(path);
      toast.success(`Storage scan initiated for ${path}`);
    } catch (error) {
      toast.error(`Failed to scan storage: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleUpdateUser = async (id: string, updates: Partial<User>) => {
    try {
      await adminApi.updateUser(id, updates);
      toast.success('User updated successfully');
    } catch (error) {
      toast.error(`Failed to update user: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <AdminPanel
        systemInfo={systemInfo || {
          version: '1.0.0',
          uptime: 86400,
          cpuUsage: 45,
          memoryUsage: 62,
          diskUsage: {
            total: 1073741824000,
            used: 536870912000,
            free: 536870912000
          },
          activeConnections: 12,
          totalRequests: 15420
        }}
        users={users || []}
        storageInfo={storageInfo || []}
        backups={backups || []}
        onCreateBackup={handleCreateBackup}
        onRestoreBackup={handleRestoreBackup}
        onScanStorage={handleScanStorage}
        onUpdateUser={handleUpdateUser}
      />
    </div>
  );
};