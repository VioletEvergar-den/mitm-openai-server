import { message } from 'antd';

/**
 * 复制内容到剪贴板
 * @param text 要复制的文本或对象
 * @returns Promise<boolean> 复制是否成功
 */
export const copyToClipboard = async (text: any): Promise<boolean> => {
  let textToCopy: string;

  // 处理各种数据类型
  if (typeof text === 'string') {
    textToCopy = text;
  } else if (text === null || text === undefined) {
    textToCopy = '';
  } else {
    try {
      textToCopy = JSON.stringify(text, null, 2);
    } catch (error) {
      textToCopy = String(text);
    }
  }

  // 检查是否有内容可复制
  if (!textToCopy) {
    message.warning('没有内容可复制');
    return false;
  }

  try {
    // 优先使用现代 Clipboard API
    if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
      await navigator.clipboard.writeText(textToCopy);
      message.success('已复制到剪贴板');
      return true;
    } else {
      // 降级方案：使用 textarea + execCommand
      const textArea = document.createElement('textarea');
      textArea.value = textToCopy;
      textArea.style.position = 'fixed';
      textArea.style.left = '-9999px';
      textArea.style.top = '-9999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      const successful = document.execCommand('copy');
      document.body.removeChild(textArea);

      if (successful) {
        message.success('已复制到剪贴板');
        return true;
      } else {
        message.error('复制失败，请手动复制');
        return false;
      }
    }
  } catch (err) {
    console.error('复制失败:', err);
    message.error('复制失败，请手动复制');
    return false;
  }
};
