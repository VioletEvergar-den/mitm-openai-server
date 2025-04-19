/**
 * 获取请求方法对应的颜色
 */
export const getMethodColor = (method: string): string => {
  const methodMap: Record<string, string> = {
    GET: 'green',
    POST: 'blue',
    PUT: 'orange',
    DELETE: 'red',
    PATCH: 'purple',
  };
  return methodMap[method] || 'default';
};

/**
 * 获取状态码对应的颜色
 */
export const getStatusColor = (status: number): string => {
  if (status >= 200 && status < 300) return 'success';
  if (status >= 300 && status < 400) return 'processing';
  if (status >= 400 && status < 500) return 'warning';
  if (status >= 500) return 'error';
  return 'default';
};

/**
 * JSON的主题样式
 */
export const jsonPrettyTheme = {
  main: 'line-height:1.3;color:#66d9ef;background:#272822;overflow:auto;padding:12px;border-radius:4px;',
  key: 'color:#f92672;',
  string: 'color:#a6e22e;',
  value: 'color:#fd971f;',
  boolean: 'color:#ac81fe;',
}; 