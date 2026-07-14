import { notification } from 'antd';
import type { ArgsProps } from 'antd/es/notification';

type NotificationType = 'success' | 'info' | 'warning' | 'error';

export const useNotification = () => {
  const [api, contextHolder] = notification.useNotification();

  const showNotification = (
    type: NotificationType,
    message: string,
    description?: string,
    config?: Omit<ArgsProps, 'message' | 'description'>
  ) => {
    api[type]({
      message,
      description,
      placement: 'topRight',
      duration: 4,
      ...config,
    });
  };

  return {
    contextHolder,
    success: (message: string, description?: string) =>
      showNotification('success', message, description),
    error: (message: string, description?: string) =>
      showNotification('error', message, description),
    info: (message: string, description?: string) =>
      showNotification('info', message, description),
    warning: (message: string, description?: string) =>
      showNotification('warning', message, description),
  };
};
