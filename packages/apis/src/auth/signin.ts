import md5 from 'md5';
import moment from 'moment';
import {
    useErrorResponseData,
    useJWT,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { Session, User } from '@nao-todo-server/models';
import type { Request, Response } from 'express';

const signin = async (req: Request, res: Response) => {
    try {
        // 判断邮箱和密码是否为空
        if (!req.body.email || !req.body.password) {
            res.json(useErrorResponseData('邮箱或密码不能为空'));
            return;
        }

        // 根据邮箱和加密后的密码查找用户
        const targetUser = await User.findOne({
            email: req.body.email,
            password: md5(req.body.password)
        });

        // 判断用户是否存在
        if (!targetUser) {
            res.json(useErrorResponseData('认证信息有误'));
            return;
        }

        // 用户存在则生成新的 Token
        const stringifyUserId = targetUser._id.toString();
        const jsonWebToken = useJWT({
            id: stringifyUserId,
            userId: stringifyUserId,
            email: targetUser.email,
            nickname: targetUser.nickname,
            avatar: targetUser.avatar,
            role: targetUser.role,
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-expect-error
            createdAt: targetUser.toJSON().createdAt,
            timeStamp: Date.now()
        });

        // 查找并更新现存 Session
        const expiresAt = moment().add(7, 'd').toDate();
        const currentSession = await Session.findOneAndUpdate(
            { userId: targetUser._id },
            {
                $set: {
                    token: jsonWebToken,
                    expiresAt: expiresAt
                }
            },
            { new: true }
        );

        // 判断更新结果
        if (currentSession) {
            res.json(useSuccessfulResponseData({ token: jsonWebToken }));
            return;
        }

        // 若不存在对应 Session，则创建新的 Session
        const newSession = await Session.create({
            userId: targetUser._id,
            token: jsonWebToken,
            expiresAt: expiresAt
        });

        // 判断创建结果
        if (newSession) {
            res.json(useSuccessfulResponseData({ token: jsonWebToken }));
        } else {
            res.json(useErrorResponseData('登录失败，请稍后重试'));
        }
    } catch (error) {
        console.log('[api/auth/signin] \n -->', error);
        res.json(useResponseData(50001, '登录失败', null));
    }
};

// 导出登录函数
export default signin;
