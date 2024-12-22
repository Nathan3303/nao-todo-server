// 导入Session模型
import { Session } from '@nao-todo-server/models';
// 导入自定义的钩子函数
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
// 导入express的Request和Response类型
import type { Request, Response } from 'express';

// 定义退出登录的异步函数
const signout = async (req: Request, res: Response) => {
    try {
        // 如果请求中没有jwt参数，抛出错误
        if (!req.query.jwt) {
            res.json(useErrorResponseData('用户凭证不能为空'));
            return;
        }

        // 解析 jwt，获取 payload 参数
        const jwt = req.query.jwt as string;
        const jwtPayload = getJWTPayload(jwt);

        // 在Session模型中查找并删除对应的session
        const session = await Session.findOneAndDelete({
            userId: jwtPayload.userId,
            token: jwt
        }).exec();

        // 如果找不到对应的session，抛出错误
        // if (!session) {
        //     res.json(useErrorResponseData('用户凭证无效'));
        //     return;
        // }

        // 返回成功响应数据
        res.json(
            useSuccessfulResponseData({
                userId: session?.userId.toString() ?? null
            })
        );
    } catch (error) {
        console.log('[api/auth/signout] \n -->', error);
        res.json(useResponseData(50001, '登出失败', null));
    }
};

// 导出signout函数
export default signout;
