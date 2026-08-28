// 실제 코드를 실행하는 진입점입니다.
import { db } from './database.js';
import { createUserService } from './user-service.js';

export const userService = createUserService({ db });