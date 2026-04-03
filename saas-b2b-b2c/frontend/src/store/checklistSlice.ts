// src/store/checklistSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import apiClient from '../api/axiosClient';
import { AxiosError } from 'axios';

// === Локальные Типы ===
export interface Task {
  id: string;
  title: string;
  description?: string;
  status: 'pending' | 'in_progress' | 'completed';
  order: number;
  created_at: string;
  updated_at: string;
  // Добавляем возможность назначать задачу (опционально)
  assigned_to?: string; 
}

export interface Checklist {
  id: string;
  title: string;
  description?: string;
  user_id: string;
  tenant_id: string;
  status: 'pending' | 'in_progress' | 'completed';
  created_at: string;
  updated_at: string;
  tasks: Task[];
  kpi_score: number;
}

interface ChecklistState {
  items: Checklist[]; // ЯВНО УКАЗЫВАЕМ, ЧТО ЭТО ВСЕГДА МАССИВ
  currentChecklist: Checklist | null;
  loading: boolean;
  error: string | null;
}

interface ApiError {
  message: string;
}

// Вспомогательная функция для обработки ошибок
const handleAsyncError = (error: unknown, defaultMessage: string): ApiError => {
  if (error instanceof AxiosError) {
    return { message: error.response?.data?.message || defaultMessage };
  }
  return { message: defaultMessage };
};

// === Async Thunks ===

export const fetchChecklists = createAsyncThunk<Checklist[], void, { rejectValue: ApiError }>(
  'checklist/fetchChecklists',
  async (_, { rejectWithValue }) => {
    try {
      const response = await apiClient.get<Checklist[]>('/api/v1/checklists');
      return response.data;
    } catch (error) {
      return rejectWithValue(handleAsyncError(error, 'Ошибка получения чек-листов'));
    }
  }
);

export const fetchChecklistById = createAsyncThunk<Checklist, string, { rejectValue: ApiError }>(
  'checklist/fetchChecklistById',
  async (id, { rejectWithValue }) => {
    try {
      const response = await apiClient.get<Checklist>(`/api/v1/checklists/${id}`);
      return response.data;
    } catch (error) {
      return rejectWithValue(handleAsyncError(error, 'Ошибка получения чек-листа'));
    }
  }
);

// Принимаем Partial<Checklist> для гибкости (чтобы передавать tasks)
export const createChecklist = createAsyncThunk<Checklist, Partial<Checklist>, { rejectValue: ApiError }>(
  'checklist/createChecklist',
  async (checklistData, { rejectWithValue }) => {
    try {
      // Бэкенд должен сам добавить id, user_id, tenant_id, created_at и т.д.
      const response = await apiClient.post<Checklist>('/api/v1/checklists', checklistData);
      return response.data;
    } catch (error) {
      return rejectWithValue(handleAsyncError(error, 'Ошибка создания чек-листа'));
    }
  }
);

export const updateChecklist = createAsyncThunk<Checklist, { id: string } & Partial<Checklist>, { rejectValue: ApiError }>(
  'checklist/updateChecklist',
  async ({ id, ...updateData }, { rejectWithValue }) => {
    try {
      const response = await apiClient.put<Checklist>(`/api/v1/checklists/${id}`, updateData);
      return response.data;
    } catch (error) {
      return rejectWithValue(handleAsyncError(error, 'Ошибка обновления чек-листа'));
    }
  }
);

export const deleteChecklist = createAsyncThunk<string, string, { rejectValue: ApiError }>(
  'checklist/deleteChecklist',
  async (id, { rejectWithValue }) => {
    try {
      await apiClient.delete(`/api/v1/checklists/${id}`);
      return id;
    } catch (error) {
      return rejectWithValue(handleAsyncError(error, 'Ошибка удаления чек-листа'));
    }
  }
);

export const completeChecklist = createAsyncThunk<Checklist, string, { rejectValue: ApiError }>(
  'checklist/completeChecklist',
  async (id, { rejectWithValue }) => {
    try {
      const response = await apiClient.post<Checklist>(`/api/v1/checklists/${id}/complete`, {});
      return response.data;
    } catch (error) {
      return rejectWithValue(handleAsyncError(error, 'Ошибка завершения чек-листа'));
    }
  }
);

// === Начальное состояние ===
const initialState: ChecklistState = {
  items: [], // Гарантируем пустой массив при старте
  currentChecklist: null,
  loading: false,
  error: null,
};

// === Slice ===
const checklistSlice = createSlice({
  name: 'checklist',
  initialState,
  reducers: {
    setCurrentChecklist: (state, action: PayloadAction<Checklist>) => {
      state.currentChecklist = action.payload;
    },
    resetCurrentChecklist: (state) => {
      state.currentChecklist = null;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch all
      .addCase(fetchChecklists.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchChecklists.fulfilled, (state, action) => {
        state.loading = false;
        // Убеждаемся, что payload - это массив
        state.items = Array.isArray(action.payload) ? action.payload : [];
      })
      .addCase(fetchChecklists.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.message || 'Ошибка загрузки';
      })
      
      // Fetch by ID
      .addCase(fetchChecklistById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchChecklistById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentChecklist = action.payload;
        const index = state.items.findIndex(item => item.id === action.payload.id);
        if (index !== -1) state.items[index] = action.payload;
      })
      .addCase(fetchChecklistById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.message || 'Ошибка загрузки';
      })
      
      // Create - ЗДЕСЬ ИСПРАВЛЕНИЕ ОШИБКИ PUSH
      .addCase(createChecklist.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createChecklist.fulfilled, (state, action) => {
        state.loading = false;
        // === ГАРАНТИЯ БЕЗОПАСНОСТИ ===
        // Если по какой-то причине items стал null/undefined, восстанавливаем его как []
        if (!state.items) {
          state.items = [];
        }
        state.items.push(action.payload);
      })
      .addCase(createChecklist.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.message || 'Ошибка создания';
      })
      
      // Update
      .addCase(updateChecklist.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateChecklist.fulfilled, (state, action) => {
        state.loading = false;
        const index = state.items.findIndex((item) => item.id === action.payload.id);
        if (index !== -1) state.items[index] = action.payload;
        if (state.currentChecklist?.id === action.payload.id) {
          state.currentChecklist = action.payload;
        }
      })
      .addCase(updateChecklist.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.message || 'Ошибка обновления';
      })
      
      // Delete
      .addCase(deleteChecklist.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteChecklist.fulfilled, (state, action) => {
        state.loading = false;
        state.items = state.items.filter((item) => item.id !== action.payload);
        if (state.currentChecklist?.id === action.payload) {
          state.currentChecklist = null;
        }
      })
      .addCase(deleteChecklist.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.message || 'Ошибка удаления';
      })
      
      // Complete
      .addCase(completeChecklist.pending, (state) => {
        state.loading = true;
      })
      .addCase(completeChecklist.fulfilled, (state, action) => {
        state.loading = false;
        const index = state.items.findIndex((item) => item.id === action.payload.id);
        if (index !== -1) state.items[index] = action.payload;
        if (state.currentChecklist?.id === action.payload.id) {
          state.currentChecklist = action.payload;
        }
      })
      .addCase(completeChecklist.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload?.message || 'Ошибка завершения';
      });
  },
});

export const { setCurrentChecklist, resetCurrentChecklist, clearError } = checklistSlice.actions;
export default checklistSlice.reducer;