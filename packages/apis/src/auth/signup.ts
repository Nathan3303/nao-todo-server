import validator from 'validator';
import {
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';
import { User } from '@nao-todo-server/models';
import md5 from 'md5';

// 注册函数
const signup = async (req: Request, res: Response) => {
    try {
        // 验证邮箱和密码是否为空
        if (!req.body.email || !req.body.password) {
            res.json(useErrorResponseData('邮箱或密码不能为空'));
            return;
        }

        // 验证邮箱格式
        if (!validator.isEmail(req.body.email)) {
            res.json(useErrorResponseData('邮箱格式不正确'));
            return;
        }

        // 根据邮箱和密码查找用户
        const targetUser = await User.findOne({
            email: req.body.email,
            password: md5(req.body.password)
        });

        // 判断对应用户是否存在
        if (targetUser) {
            res.json(useErrorResponseData('邮箱已被注册'));
            return;
        }

        // 不存在，则创建用户
        const newUser = await User.create({
            account: req.body.email,
            password: md5(req.body.password),
            email: req.body.email,
            nickname: req.body.nickname,
            role: 'user'
        });

        // 判断是否创建成功
        if (newUser) {
            res.json(useSuccessfulResponseData('注册成功'));
        } else {
            res.json(useErrorResponseData('注册失败，请稍后重试'));
        }
    } catch (error) {
        console.log(
            '[api/auth/signup]',
            error instanceof Error ? error.message : error
        );
        res.json(useResponseData(50001, '注册失败', null));
    }
};

export default signup; // 导出注册函数
