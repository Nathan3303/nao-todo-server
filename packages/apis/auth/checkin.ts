import moment from 'moment';
import { User, Session } from '@nao-todo-server/models';
import {
    useJWT,
    verifyJWT,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const checkin = async (req: Request, res: Response) => {
    try {
        if (req.method !== 'GET') throw new Error('请求无效');

        if (!req.query.jwt) throw new Error('用户凭证无效，请重新登录');
        const { jwt } = req.query;
        const isJWTValid = verifyJWT(jwt as string);
        if (!isJWTValid) throw new Error('用户凭证无效，请重新登录');

        const session = await Session.findOne({ token: jwt });
        if (!session) throw new Error('用户凭证无效，请重新登录');

        if (moment().isAfter(session.expiresAt)) {
            await Session.deleteOne({ _id: session._id });
            throw new Error('用户凭证已过期，请重新登录');
        }

        const user = await User.findOne({
            _id: new ObjectId(session.userId)
        }).select({
            _id: 0,
            id: { $toString: '$_id' },
            name: 1,
            email: 1,
            nickName: 1,
            avatar: 1
        });

        if (!user) throw new Error('用户凭证无效，请重新登录');

        const newJWT = useJWT(user);
        await Session.updateOne({ _id: session._id }, { token: newJWT.jwt });

        res.json(useSuccessfulResponseData(newJWT.jwt));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData((e as Error).message));
        }
        console.log('[api/auth/checkin] Unknown error:', e);
    }
};

export default checkin;
