import { db } from './database.js';
export function getUser(id) { return db.query('...', id); }