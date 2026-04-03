import apiClient from './axiosClient';
import { Lead, CreateLeadRequest, LeadActivity } from '../types';

export const getLeads = async (): Promise<Lead[]> => {
  const response = await apiClient.get<Lead[]>('/leads');
  return response.data;
};

export const createLead = async (data: CreateLeadRequest): Promise<Lead> => {
  const response = await apiClient.post<Lead>('/leads', data);
  return response.data;
};

export const updateLeadStatus = async (id: string, status: string): Promise<void> => {
  await apiClient.put(`/leads/${id}/status`, { status });
};

export const addLeadActivity = async (leadId: string, type: string, description: string): Promise<void> => {
  await apiClient.post(`/leads/${leadId}/activities`, { type, description });
};

export const getLeadDetails = async (id: string): Promise<{ lead: Lead; activities: LeadActivity[] }> => {
  const response = await apiClient.get(`/leads/${id}`);
  return response.data;
};