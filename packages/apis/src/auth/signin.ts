import md5 from 'md5';
import moment from 'moment';
import {
    useErrorResponseData,
    useJWT,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { User, Session } from '@nao-todo-server/models';
import type { Request, Response } from 'express';

const findUser = async (email: string, password: string) => {
    try {
        const user = await User.findOne({
            email,
            password: md5(password)
        }).exec();
        return user;
    } catch (e: unknown) {
        console.log('[api/auth/signin] _findUser error:', e);
    }
};

const signin = async (req: Request, res: Response) => {
    try {
        if (!req.body.email || !req.body.password)
            throw new Error('邮箱或密码不能为空');

        const { email, password } = req.body;
        const user = await findUser(email, password);
        if (!user) throw new Error('邮箱或密码错误');

        const userId = user._id.toString();
        const expiresAt = moment().add(7, 'days').utcOffset(8).utc().format();

        const token = useJWT({
            userId,
            email: user.email,
            nickName: user.nickName,
            expiresAt
        });

        let session = await Session.findOneAndUpdate(
            { userId },
            { $set: { token, expiresAt } }
        );

        if (!session) {
            session = await Session.create({
                userId,
                token,
                expiresAt
            });
        }

        if (!session) throw new Error('登录失败，请稍后重试');

        res.json(useSuccessfulResponseData({ token }));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/auth/signin] Error:', e);
    }
};

export default signin;
export const findUserByEmailAndPassword = findUser;
