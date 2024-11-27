import { Comment } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const updateComment = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.query.commentId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 更新数据
        const updateResult = await Comment.updateOne(
            {
                _id: new ObjectId(req.query.commentId as string),
                userId: new ObjectId(userId)
            },
            {
                $set: {
                    content: req.body.content
                }
            }
        ).exec();

        // 判断更新结果
        if (updateResult && updateResult.modifiedCount) {
            res.json(
                useSuccessfulResponseData({
                    commentId: req.query.commentId
                })
            );
        } else {
            return res.json(res.json(useErrorResponseData('更新失败')));
        }
    } catch (e: unknown) {
        console.log('[api/comment/updateComment] Error:', e);
        return res.json(useResponseData(50001, '更新失败，请稍后再试', null));
    }
};

const updateComments = async () => {};

export { updateComment, updateComments };
