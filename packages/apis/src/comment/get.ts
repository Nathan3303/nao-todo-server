import { Comment } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const getComment = async (req: Request, res: Response) => {
    res.json(useSuccessfulResponseData('Hello World'));
};

const getComments = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.query.todoId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 获取数据
        // const comments = await Comment.find({
        //     userId: new ObjectId(userId),
        //     todoId: new ObjectId(req.query.todoId as string),
        //     commentId: null
        // }).exec();
        const comments = await Comment.aggregate([
            {
                $match: {
                    userId: new ObjectId(userId),
                    todoId: new ObjectId(req.query.todoId as string),
                    commentId: null
                }
            },
            {
                $lookup: {
                    from: 'users',
                    localField: 'userId',
                    foreignField: '_id',
                    as: 'user'
                }
            },
            {
                $unwind: {
                    path: '$user',
                    preserveNullAndEmptyArrays: true
                }
            },
            {
                $project: {
                    _id: 0,
                    id: '$_id',
                    commentId: 1,
                    content: 1,
                    createdAt: 1,
                    attachments: 1,
                    isTopUp: 1,
                    user: {
                        avatar: 1,
                        nickname: 1
                    }
                }
            }
        ]);

        // 返回数据
        res.json(useSuccessfulResponseData(comments));
    } catch (e: unknown) {
        console.log('[api/comment/getComments] Error:', e);
        return res.json(useResponseData(50001, '获取失败，请稍后再试', null));
    }
};

export { getComment, getComments };
