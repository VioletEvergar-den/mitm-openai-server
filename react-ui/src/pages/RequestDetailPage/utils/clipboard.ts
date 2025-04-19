import { message } from 'antd';

/**
 * 复制内容到剪贴板
 * @param text 要复制的文本或对象
 */
export const copyToClipboard = (text: any) => {
  if (typeof text !== 'string') {
    text = JSON.stringify(text, null, 2);
  }
  
  navigator.clipboard.writeText(text)
    .then(() => {
      message.success('已复制到剪贴板');
    })
    .catch(err => {
      message.error('复制失败');
      console.error('复制失败:', err);
    });
}; 