import moment from 'moment';
import { Session, User } from '@nao-todo-server/models';
import {
    useErrorResponseData,
    useJWT,
    useResponseData,
    useSuccessfulResponseData,
    verifyJWT
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

// 检查用户是否登录
const checkin = async (req: Request, res: Response) => {
    try {
        // 检查请求中是否包含 jwt
        if (!req.query.jwt) {
            res.json(useErrorResponseData('用户凭证无效，请重新登录'));
            return;
        }

        // 验证 jwt 是否有效
        const jwt = req.query.jwt as string;
        const isJWTValid = verifyJWT(jwt);
        if (!isJWTValid) {
            res.json(useErrorResponseData('用户凭证无效，请重新登录'));
            return;
        }

        // 根据 jwt 查找 session
        const session = await Session.findOne({ token: jwt }).exec();
        if (!session) {
            res.json(useErrorResponseData('用户凭证无效，请重新登录'));
            return;
        }

        // 检查 session 是否过期
        const isExpired = moment().isAfter(session.expiresAt);
        if (isExpired) {
            await Session.deleteOne({ _id: session._id }).exec();
            res.json(useErrorResponseData('用户凭证已过期，请重新登录'));
            return;
        }

        // 根据 session 中的 userId 查找用户
        const user = await User.findOne({
            _id: session.userId
        }).exec();

        // 判断用户是否存在
        if (!user) {
            await Session.deleteOne({ _id: session._id }).exec();
            res.json(useErrorResponseData('用户凭证无效，请重新登录'));
            return;
        }

        // 生成新的 jwt
        const stringifyUserId = user._id.toString(); // 获取用户ID
        const jsonWebToken = useJWT({
            id: stringifyUserId,
            userId: stringifyUserId,
            email: user.email,
            nickname: user.nickname,
            avatar: user.avatar,
            role: user.role,
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-expect-error
            createdAt: user.toJSON().createdAt,
            timeStamp: Date.now()
        });

        // 更新 session 中的 Token
        const updateResult = await Session.updateOne(
            { _id: session._id },
            { token: jsonWebToken },
            { new: true }
        ).exec();

        // 判断更新结果
        if (updateResult) {
            res.json(useSuccessfulResponseData({ token: jsonWebToken }));
        } else {
            res.json(useErrorResponseData('用户凭证无效，请重新登录'));
        }
    } catch (error) {
        console.log('[api/auth/checkin] \n -->', error);
        res.json(useResponseData(50001, '用户凭证无效，请重新登录', null));
    }
};

export default checkin;
