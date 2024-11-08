import md5 from 'md5';
import validator from 'validator';
import { User } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
// import { findUserByEmailAndPassword } from './signin';
import type { Request, Response } from 'express';

// const _createUser = async (email: string, password: string) => {
//     try {
//         const user = await User.create({
//             username: email,
//             password: md5(password),
//             email,
//             nickName: `${email.split('@')[0]}`
//         });
//         return user;
//     } catch (e) {
//         console.log('[api/auth/signup] _createUser Error:', e);
//         return null;
//     }
// };

const signup = async (req: Request, res: Response) => {
    // try {
    //     if (!req.body.email || !req.body.password)
    //         throw new Error('邮箱或密码不能为空');

    //     const { email, password } = req.body;
    //     if (!validator.isEmail(email)) throw new Error('邮箱格式不正确');
    //     const existUser = await findUserByEmailAndPassword(email, password);
    //     if (existUser) throw new Error('邮箱已被注册');

    //     const createResult = await _createUser(email, password);
    //     if (!createResult) throw new Error('注册失败');

    //     res.json(useSuccessfulResponseData('注册成功'));
    // } catch (e: unknown) {
    //     if (e instanceof Error) {
    //         res.json(useErrorResponseData(e.message));
    //     }
    //     console.log('[api/auth/signup] Error:', e);
    // }
};

export default signup;
// export const createUserByEmailAndPassword = _createUser;
