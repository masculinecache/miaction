import { Bill, CommitteeMeeting, Representative, EmailTemplate, MailtoLink, Location } from '../types'

const API_URL = (import.meta as any).env?.VITE_API_URL || ''

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options,
  })
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status} ${response.statusText}`)
  }
  
  return response.json()
}

export const api = {
  getBills: () => fetchAPI<Bill[]>('/api/bills'),
  getScheduledBills: () => fetchAPI<Bill[]>('/api/bills/scheduled'),
  getBill: (id: string) => fetchAPI<Bill>(`/api/bills/${id}`),
  getMeetings: () => fetchAPI<CommitteeMeeting[]>('/api/meetings'),
  
  findRepresentatives: (location: Location) => 
    fetchAPI<Representative[]>('/api/representatives', {
      method: 'POST',
      body: JSON.stringify(location),
    }),
  
  searchRepresentatives: (query: string) => 
    fetchAPI<Representative[]>(`/api/representatives/search?q=${encodeURIComponent(query)}`),
  
  getEmailTemplates: () => fetchAPI<EmailTemplate[]>('/api/email/templates'),
  
  composeEmail: (templateId: string, variables: Record<string, string>, repIds: string[]) => 
    fetchAPI<MailtoLink[]>('/api/email/compose', {
      method: 'POST',
      body: JSON.stringify({
        templateId,
        variables,
        representativeIds: repIds,
        senderName: variables.SenderName || '',
        senderEmail: variables.SenderEmail || '',
      }),
    }),
  
  previewEmail: (templateId: string, variables: Record<string, string>) => 
    fetchAPI<{ subject: string; body: string }>('/api/email/preview', {
      method: 'POST',
      body: JSON.stringify({ templateId, variables }),
    }),
}
