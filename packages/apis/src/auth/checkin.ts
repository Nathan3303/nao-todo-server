import moment from 'moment';
import { User, Session } from '@nao-todo-server/models';
import {
    useJWT,
    useErrorResponseData,
    useSuccessfulResponseData,
    verifyJWT
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

// 检查用户是否登录
const checkin = async (req: Request, res: Response) => {
    try {
        // 检查请求中是否包含jwt
        if (!req.query.jwt) throw new Error('用户凭证无效，请重新登录');

        const jwt = req.query.jwt as string;

        // 验证jwt是否有效
        const isJWTValid = verifyJWT(jwt);
        if (!isJWTValid) throw new Error('用户凭证无效，请重新登录');

        // 根据jwt查找session
        const session = await Session.findOne({ token: jwt });
        if (!session) throw new Error('用户凭证无效，请重新登录');

        // 检查session是否过期
        if (moment().isAfter(session.expiresAt)) {
            await Session.deleteOne({ _id: session._id }).exec();
            throw new Error('用户凭证已过期，请重新登录');
        }

        // 根据session中的userId查找用户
        const user = await User.findOne({
            _id: new ObjectId(session.userId)
        }).exec();
        if (!user) throw new Error('用户凭证无效，请重新登录');

        // 生成新的jwt
        const newJWT = useJWT({
            userId: user._id,
            email: user.email,
            nickName: user.nickName,
            avatar: user.avatar,
            role: user.role,
            expiresAt: moment().add(7, 'days').utcOffset(8).utc().format()
        });

        // 更新session中的token
        await Session.updateOne({ _id: session._id }, { token: newJWT }).exec();

        // 返回新的jwt
        res.json(useSuccessfulResponseData({ token: newJWT }));
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData((e as Error).message));
        }
        console.log('[api/auth/checkin] Error:', e);
    }
};

export default checkin;
