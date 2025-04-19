import { RequestRecord } from '../types/RequestRecord';

// API基础URL
const API_BASE_URL = process.env.NODE_ENV === 'production' ? '' : 'http://localhost:8080';

export class RequestService {
  /**
   * 获取所有请求记录
   */
  static async getRequests(): Promise<RequestRecord[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/ui/api/requests`);
      if (!response.ok) {
        throw new Error(`获取请求记录失败: ${response.status}`);
      }
      const data = await response.json();
      return data.data.requests || [];
    } catch (error) {
      console.error('获取请求记录时出错:', error);
      throw error;
    }
  }

  /**
   * 获取单个请求记录详情
   */
  static async getRequestById(id: string): Promise<RequestRecord> {
    try {
      const response = await fetch(`${API_BASE_URL}/ui/api/requests/${id}`);
      if (!response.ok) {
        throw new Error(`获取请求记录详情失败: ${response.status}`);
      }
      const data = await response.json();
      return data.data || {};
    } catch (error) {
      console.error('获取请求记录详情时出错:', error);
      throw error;
    }
  }

  /**
   * 删除所有请求记录
   */
  static async deleteAllRequests(): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/ui/api/requests`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        throw new Error(`删除所有请求记录失败: ${response.status}`);
      }
    } catch (error) {
      console.error('删除所有请求记录时出错:', error);
      throw error;
    }
  }
} 