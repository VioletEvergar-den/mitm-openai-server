import { message } from 'antd';

/**
 * 将数据导出为JSON文件并下载
 * @param data 要导出的数据
 * @param filename 文件名（不包含扩展名）
 */
export const exportAsJson = (data: any, filename: string): void => {
  try {
    // 确保数据是可序列化的
    const jsonString = JSON.stringify(data, null, 2);
    
    // 创建Blob对象
    const blob = new Blob([jsonString], { type: 'application/json' });
    
    // 创建下载链接
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    
    // 设置链接属性
    link.href = url;
    link.download = `${filename}.json`;
    
    // 添加到DOM，触发下载，然后移除
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    // 释放URL对象
    URL.revokeObjectURL(url);
    
    message.success('导出成功');
  } catch (error) {
    console.error('导出失败:', error);
    message.error('导出失败');
  }
}; 