export interface RequestRecord {
  id: string;
  model: string;
  path: string;
  method: string;
  type: string;
  timestamp: string;
  response?: {
    choices?: Array<{
      message?: {
        content?: string;
      };
    }>;
  };
  [key: string]: any; // 允许其他任意字段
} 