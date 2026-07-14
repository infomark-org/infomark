import { useState, useEffect, useCallback } from 'react';
import type { AsyncState, ApiError } from '@/types';

export function useApi<T>(
  apiFunc: () => Promise<T>,
  immediate = true
): AsyncState<T> & { refetch: () => Promise<void> } {
  const [state, setState] = useState<AsyncState<T>>({
    data: null,
    loading: immediate,
    error: null,
  });

  const execute = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    
    try {
      const data = await apiFunc();
      setState({ data, loading: false, error: null });
    } catch (error) {
      setState({
        data: null,
        loading: false,
        error: error as ApiError,
      });
    }
  }, [apiFunc]);

  useEffect(() => {
    if (immediate) {
      execute();
    }
  }, [immediate, execute]);

  return {
    ...state,
    refetch: execute,
  };
}

export function useAsyncAction<T, P = void>() {
  const [state, setState] = useState<AsyncState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (apiFunc: (params: P) => Promise<T>, params: P) => {
    setState({ data: null, loading: true, error: null });
    
    try {
      const data = await apiFunc(params);
      setState({ data, loading: false, error: null });
      return data;
    } catch (error) {
      const apiError = error as ApiError;
      setState({ data: null, loading: false, error: apiError });
      throw apiError;
    }
  }, []);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return {
    ...state,
    execute,
    reset,
  };
}
