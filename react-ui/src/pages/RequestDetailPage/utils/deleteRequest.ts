import { Modal, message } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import React from 'react';
import { apiService } from '../../../services/api';

const { confirm } = Modal;

/**
 * 删除请求确认框
 * @param id 请求ID
 * @param onSuccess 删除成功回调
 */
export const confirmDeleteRequest = (id: string, onSuccess: () => void) => {
  if (!id) return;

  confirm({
    title: '确认删除',
    icon: React.createElement(ExclamationCircleOutlined),
    content: '确定要删除此请求记录吗？此操作不可撤销。',
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await apiService.deleteRequest(id);
        message.success('请求已成功删除');
        onSuccess();
      } catch (err) {
        console.error('Failed to delete request:', err);
        message.error('删除请求失败');
      }
    },
  });
}; 