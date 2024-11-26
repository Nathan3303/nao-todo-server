import md5 from 'md5';
import { User } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

// 更新用户昵称
const updateUserNickname = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.body.nickname) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 更新用户昵称
        const updateResult = await User.findOneAndUpdate(
            { _id: userId },
            { $set: { nickname: (req.body.nickname as string).trim() } },
            { new: true }
        ).exec();

        // 判断更新结果，返回对应的 JSON 数据
        if (updateResult) {
            res.json(
                useSuccessfulResponseData({
                    userId: updateResult._id.toString()
                })
            );
        } else {
            res.json(useErrorResponseData('更新用户昵称失败'));
        }
    } catch (error) {
        console.log('[api/user/updateUserNickname] Error:', error);
        res.json(useResponseData(50001, '更新用户昵称失败，请稍后再试', null));
    }
};

// 更新密码
const updateUserPassword = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.body.password) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 更新用户密码
        const updateResult = await User.findOneAndUpdate(
            { _id: userId },
            { $set: { password: md5(req.body.password) } },
            { new: true }
        );

        // 判断更新结果，返回对应的 JSON 数据
        if (updateResult) {
            res.json(
                useSuccessfulResponseData({
                    userId: updateResult._id.toString()
                })
            );
        } else {
            res.json(useErrorResponseData('密码更新失败'));
        }
    } catch (error) {
        console.log('[api/user/updatePassword] Error:', error);
        res.json(useResponseData(50001, '密码更新失败，请稍后再试', null));
    }
};

export { updateUserNickname, updateUserPassword };
