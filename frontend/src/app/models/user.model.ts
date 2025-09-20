export type UserRole = 'ANCHOR' | 'SUPPLIER';

export interface User {
    id: string;
    username: string;
    role: UserRole;
    password?: string;
}