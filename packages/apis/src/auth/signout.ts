// 导入Session模型
import { Session } from '@nao-todo-server/models';
// 导入自定义的钩子函数
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload
} from '@nao-todo-server/hooks';
// 导入express的Request和Response类型
import type { Request, Response } from 'express';

// 定义退出登录的异步函数
const signout = async (req: Request, res: Response) => {
    try {
        // 如果请求中没有jwt参数，抛出错误
        if (!req.query.jwt) throw new Error('用户凭证不能为空');

        // 获取jwt参数
        const jwt = req.query.jwt as string;

        // 解析jwt参数，获取payload
        const jwtPayload = getJWTPayload(jwt);

        // 在Session模型中查找并删除对应的session
        const session = await Session.findOneAndDelete({
            token: jwt,
            userId: jwtPayload.userId
        }).exec();

        // 如果找不到对应的session，抛出错误
        if (!session) throw new Error('用户凭证无效');

        // 返回成功响应数据
        res.json(useSuccessfulResponseData({ userId: session.userId }));
    } catch (e: unknown) {
        // 如果抛出的是Error类型的错误，返回错误响应数据
        if (e instanceof Error) {
            res.json(useErrorResponseData(e.message));
            return;
        }
        // 如果抛出的是其他类型的错误，打印错误信息，并返回错误响应数据
        console.log('[api/auth/signout] Error:', e);
        res.json(useErrorResponseData('退出失败，请稍后重试'));
    }
};

// 导出signout函数
export default signout;
