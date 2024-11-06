import md5 from 'md5';
import moment from 'moment';
import {
    useJWT,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { User, Session } from '@nao-todo-server/models';
import type { Request, Response } from 'express';

const _findUser = async (email: string, password: string) => {
    try {
        const user = await User.findOne({
            email,
            password: md5(password)
        }).select({
            // _id: 0,
            id: { $toString: '$_id' },
            name: 1,
            email: 1,
            nickName: 1,
            avatar: 1
        });
        return user;
    } catch (e: unknown) {
        console.log('[api/auth/signin] _findUser error:', e);
    }
};

const _createSession = async (user: InstanceType<typeof User>) => {
    try {
        const createdAt = moment().utcOffset(8).utc().format();
        const expiresAt = moment().add(7, 'days').utcOffset(8).utc().format();
        const sessionFromDB = await Session.findOne({ userId: user.id });
        if (sessionFromDB) {
            await Session.updateOne(
                { _id: sessionFromDB.id },
                { $set: { createdAt, expiresAt } }
            );
            return sessionFromDB.token;
        }
        const { jwt } = useJWT(user);
        const session = await Session.create({
            userId: user.id,
            token: jwt,
            createdAt,
            expiresAt
        });
        return session ? jwt : null;
    } catch (e: unknown) {
        console.log('[api/auth/signin] _createSession error:', e);
        return null;
    }
};

const signin = async (req: Request, res: Response) => {
    try {
        if (req.method !== 'POST') throw new Error('请求无效');

        if (!req.body.email || !req.body.password)
            throw new Error('邮箱或密码不能为空');

        const { email, password } = req.body;
        const user = await _findUser(email, password);
        if (!user) throw new Error('邮箱或密码错误');

        const token = await _createSession(user);
        if (!token) throw new Error('登录失败，请稍后重试');

        res.json(useSuccessfulResponseData(token));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/auth/signin] Error:', e);
    }
};

export default signin;
export const findUserByEmailAndPassword = _findUser;
