import { useEffect } from 'react';

interface NavigationKeysProps {
  prevAction: () => void;
  nextAction: () => void;
  hasPrev: boolean;
  hasNext: boolean;
}

export const useNavigationKeys = ({
  prevAction,
  nextAction,
  hasPrev,
  hasNext
}: NavigationKeysProps) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // 避免在输入框中触发
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }
    
      // 左方向键导航到上一个请求
      if (e.key === 'ArrowLeft' && hasPrev) {
        prevAction();
      }
      // 右方向键导航到下一个请求
      else if (e.key === 'ArrowRight' && hasNext) {
        nextAction();
      }
    };

    // 添加键盘事件监听
    window.addEventListener('keydown', handleKeyDown);

    // 清理函数
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [hasPrev, hasNext, prevAction, nextAction]);
};

export default useNavigationKeys; 