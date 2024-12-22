import { Session } from '@nao-todo-server/models';
import {
    useErrorResponseData,
    useResponseData,
    verifyJWT
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const validate = async (req: Request, res: Response) => {
    try {
        const { authorization } = req.headers;
        if (!authorization) {
            res.json(useResponseData(40300, '用户凭证无效', null));
            return false;
        }
        const token = authorization.replace('Bearer ', '');
        const isJWTValid = verifyJWT(token);
        if (!isJWTValid) {
            res.json(useResponseData(40300, '用户凭证无效', null));
            return false;
        }
        const session = await Session.findOne({ token }).exec();
        if (!session) {
            res.json(useResponseData(40300, '用户凭证无效', null));
            return false;
        }
        return true;
    } catch (e: unknown) {
        if (e instanceof Error) {
            res.json(useErrorResponseData((e as Error).message));
            return false;
        }
        console.log('[api/auth/checkin] Error:', e);
        return false;
    }
};

export default validate;
