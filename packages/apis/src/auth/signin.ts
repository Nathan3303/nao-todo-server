import md5 from 'md5'; // 导入md5加密库
import moment from 'moment'; // 导入moment库，用于处理日期和时间
import {
    useErrorResponseData,
    useJWT,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks'; // 导入自定义的hooks，用于处理错误响应、JWT和成功响应
import { User, Session } from '@nao-todo-server/models'; // 导入自定义的模型，用于操作数据库
import type { Request, Response } from 'express'; // 导入express的Request和Response类型

// 根据邮箱和密码查找用户
const findUser = async (email: string, password: string) => {
    try {
        // 根据邮箱和加密后的密码查找用户
        const user = await User.findOne({
            email,
            password: md5(password)
        }).exec();
        return user;
    } catch (e: unknown) {
        // 打印错误信息
        console.log('[api/auth/signin] _findUser error:', e);
    }
};

// 用户登录
const signin = async (req: Request, res: Response) => {
    try {
        // 判断邮箱和密码是否为空
        if (!req.body.email || !req.body.password)
            throw new Error('邮箱或密码不能为空');

        const { email, password } = req.body; // 获取邮箱和密码
        const user = await findUser(email, password); // 根据邮箱和密码查找用户
        if (!user) throw new Error('邮箱或密码错误'); // 如果用户不存在，抛出错误

        const userId = user._id.toString(); // 获取用户ID
        const expiresAt = moment().add(7, 'days').utcOffset(8).utc().format(); // 设置过期时间为7天后

        // 生成JWT
        const token = useJWT({
            userId,
            email: user.email,
            nickName: user.nickName,
            avatar: user.avatar,
            role: user.role,
            expiresAt
        });

        // 更新或创建session
        let session = await Session.findOneAndUpdate(
            { userId },
            { $set: { token, expiresAt } }
        ).exec();

        if (!session) {
            session = await Session.create({
                userId,
                token,
                expiresAt
            });
        }

        if (!session) throw new Error('登录失败，请稍后重试'); // 如果session不存在，抛出错误

        // 返回成功响应
        res.json(useSuccessfulResponseData({ token }));
    } catch (e: unknown) {
        // 如果错误是Error类型，返回错误响应
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        // 打印错误信息
        console.log('[api/auth/signin] Error:', e);
        return res.json(useErrorResponseData('登录失败，请稍后重试'));
    }
};

// 导出登录函数
export default signin;
// 导出根据邮箱和密码查找用户的函数
export const findUserByEmailAndPassword = findUser;
