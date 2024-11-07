import { Session } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    useResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const signout = async (req: Request, res: Response) => {
    try {
        if (!req.query.jwt) throw new Error('用户凭证不能为空');

        const { jwt } = req.query;

        const jwtPayload = (jwt as string).split('.')[1];
        const user = JSON.parse(atob(jwtPayload));

        const session = await Session.findOneAndDelete({
            token: jwt,
            userId: user.id
        });

        if (!session) throw new Error('用户凭证无效');

        res.json(
            useResponseData(20000, '退出成功', { userId: session.userId })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            res.json(useErrorResponseData(e.message));
            return;
        }
        console.log('[api/auth/signout] Error:', e);
        res.json(useErrorResponseData('退出失败，请稍后重试'));
    }
};

export default signout;
