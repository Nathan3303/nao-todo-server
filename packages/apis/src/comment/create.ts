import { Comment } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const createComment = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.body.todoId || !req.body.content) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 创建数据
        const createdComment = await Comment.create({
            userId: new ObjectId(userId),
            todoId: new ObjectId(req.body.todoId as string),
            content: req.body.content
        });

        // 判断创建结果
        if (createdComment) {
            res.json(
                useSuccessfulResponseData({
                    ...createdComment.toJSON(),
                    id: createdComment._id.toString()
                })
            );
        } else {
            res.json(useErrorResponseData('创建失败'));
        }
    } catch (e: unknown) {
        console.log('[api/comment/createComment] Error:', e);
        res.json(useResponseData(50001, '创建失败，请稍后再试', null));
    }
};

const createComments = async () => {};

export { createComment, createComments };
