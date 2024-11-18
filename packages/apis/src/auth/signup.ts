import md5 from 'md5'; // 导入md5加密库
import validator from 'validator'; // 导入validator库，用于验证邮箱格式
import { User } from '@nao-todo-server/models'; // 导入User模型
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks'; // 导入自定义的响应数据钩子
import { findUserByEmailAndPassword } from './signin'; // 导入根据邮箱和密码查找用户的函数
import type { Request, Response } from 'express'; // 导入express的Request和Response类型

// 创建用户函数
const createUser = async (email: string, password: string) => {
    try {
        // 创建用户，密码使用md5加密
        const user = await User.create({
            account: email,
            password: md5(password),
            email,
            nickName: `${email.split('@')[0]}` // 用户昵称为邮箱前缀
        });
        return user;
    } catch (e) {
        console.log('[api/auth/signup] _createUser Error:', e); // 打印错误日志
        return null;
    }
};

// 注册函数
const signup = async (req: Request, res: Response) => {
    try {
        // 验证邮箱和密码是否为空
        if (!req.body.email || !req.body.password)
            throw new Error('邮箱或密码不能为空');

        const { email, password } = req.body;
        // 验证邮箱格式
        if (!validator.isEmail(email)) throw new Error('邮箱格式不正确');
        // 根据邮箱和密码查找用户
        const existUser = await findUserByEmailAndPassword(email, password);
        // 如果用户已存在，抛出错误
        if (existUser) throw new Error('邮箱已被注册');

        // 创建用户
        const createResult = await createUser(email, password);
        // 如果创建失败，抛出错误
        if (!createResult) throw new Error('注册失败');

        // 返回注册成功的响应数据
        res.json(useSuccessfulResponseData('注册成功'));
    } catch (e: unknown) {
        // 如果抛出错误，返回错误响应数据
        if (e instanceof Error) {
            res.json(useErrorResponseData(e.message));
        }
        console.log('[api/auth/signup] Error:', e); // 打印错误日志
    }
};

export default signup; // 导出注册函数
export const createUserByEmailAndPassword = createUser; // 导出创建用户函数
