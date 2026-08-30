export interface Bill {
  id: string;
  number: string;
  title: string;
  description: string;
  subject?: string;
  chamber: 'house' | 'senate';
  status: string;
  sponsor: string;
  coSponsors: string[];
  committee?: string;
  introducedDate: string;
  lastAction: string;
  lastActionDate: string;
  scheduledVote?: ScheduledVote;
  url: string;
  analysisDocuments?: AnalysisDocument[];
  billDocuments?: BillDocument[];
}

export interface AnalysisDocument {
  title: string;
  description?: string;
}

export interface BillDocument {
  title: string;
  description?: string;
  pdfUrl?: string;
  htmlUrl?: string;
}

export interface ScheduledVote {
  date: string;
  body: string;
  voteType: string;
  description: string;
}

export interface CommitteeMeeting {
  committee: string;
  date: string;
  time: string;
  location: string;
  bills: string[];
  chamber: 'house' | 'senate';
  url: string;
}

export interface Representative {
  id: string;
  name: string;
  chamber: string;
  district: number;
  party: string;
  email: string;
  phone: string;
  office: string;
  url: string;
  photoUrl?: string;
}

export interface Location {
  street: string;
  city: string;
  state: string;
  zip: string;
}

export interface EmailTemplate {
  id: string;
  name: string;
  subject: string;
  description: string;
  body: string;
  variables: TemplateVariable[];
}

export interface TemplateVariable {
  name: string;
  description: string;
  required: boolean;
  default?: string;
}

export interface EmailRequest {
  templateId: string;
  variables: Record<string, string>;
  representativeIds: string[];
  senderName: string;
  senderEmail: string;
}

export interface MailtoLink {
  repId: string;
  name: string;
  mailto: string;
}
