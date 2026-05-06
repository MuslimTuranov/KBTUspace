export type Role = 'student' | 'organizer' | 'admin';
export type Scope = 'faculty' | 'global';
export type PostStatus = 'draft' | 'pending' | 'approved' | 'rejected';
export type ReportStatus = 'pending' | 'closed' | 'rejected';
export type ReportTargetType = 'post' | 'event';

export interface User {
  id: number; email: string; role: Role; faculty_id: number | null; is_banned: boolean; created_at: string; updated_at: string;
}
export interface Post {
  id: number; author_id: number; faculty_id: number | null; title: string; content: string; image_url: string | null;
  is_pinned: boolean; scope: Scope; status: PostStatus; approved_by: number | null; approved_at: string | null;
  rejection_reason: string | null; created_at: string; updated_at: string;
}
export interface Event {
  id: number; author_id: number; faculty_id: number | null; title: string; description: string; image_url: string | null;
  event_date: string; location: string; capacity: number; current_count: number; is_pinned: boolean;
  scope: Scope; status: PostStatus; approved_by: number | null; approved_at: string | null;
  rejection_reason: string | null; created_at: string; updated_at: string;
}
export interface Report {
  id: number; reporter_id: number; target_post_id: number; target_type: ReportTargetType; reason: string;
  status: ReportStatus; review_note: string | null; reviewed_by: number | null; reviewed_at: string | null;
  target_title: string; target_author_id: number; created_at: string; updated_at: string;
}
export interface CreatePostRequest { title: string; content: string; image_url?: string; scope?: Scope; faculty_id?: number; }
export interface CreateEventRequest { title: string; description: string; event_date: string; location: string; capacity: number; image_url?: string; scope?: Scope; faculty_id?: number; }
export interface CreateReportRequest { target_type: ReportTargetType; target_id: number; reason: string; }
export interface PendingContent { posts: Post[]; events: Event[]; }
export interface AuthTokenPayload { user_id: number; role: Role; faculty_id: number | null; exp: number; }
