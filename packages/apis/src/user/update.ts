import md5 from 'md5';
import { User } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
// import { calculateFileHashSync } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

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

// 更新用户头像
const updateUserAvatar = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!req.file) {
            return res.json(useErrorResponseData('未检测到文件'));
        }

        // 计算文件哈希
        // const hash = calculateFileHashSync(req.file.buffer);

        // 生成最终文件名
        const ext = path.extname(req.file.originalname);
        const filename = `avatar.${userId}${ext}`;
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        const __dirname = path.dirname(fileURLToPath(import.meta.url));
        const uploadDir = path.join(__dirname, 'avatars');
        const uploadPath = path.join(uploadDir, filename);

        // 保存文件
        await fs.promises.appendFile(uploadPath, req.file.buffer);

        // 定义访问 URL
        const url = PROD
            ? `https://todo.nathan33.site:3002/statics/avatars/${filename}`
            : `http://localhost:3002/statics/avatars/${filename}`;

        // 更新 user.avatar
        const updateRes = await User.findOneAndUpdate(
            { _id: userId },
            { $set: { avatar: url } },
            { new: true }
        );

        // 返回 URL
        if (updateRes) {
            res.json(useSuccessfulResponseData({ url }));
        } else {
            res.json(useErrorResponseData('更新用户信息失败'));
        }
    } catch (error) {
        console.log('[api/user/updateUserAvatar] Error:', error);
        res.json(useResponseData(50001, '头像更新失败，请稍后再试', null));
    }
};

export { updateUserNickname, updateUserPassword, updateUserAvatar };
