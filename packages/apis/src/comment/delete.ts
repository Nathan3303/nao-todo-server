import { Comment } from '@nao-todo-server/models';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const deleteComment = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的 UserId
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否合法
        if (!userId || !req.query.commentId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
            return;
        }

        // 删除数据
        const deleteResult = await Comment.deleteOne({
            _id: new ObjectId(req.query.commentId as string),
            userId: new ObjectId(userId)
        }).exec();

        // 判断是否删除成功
        if (deleteResult && deleteResult.deletedCount) {
            res.json(
                useSuccessfulResponseData({ commentId: req.query.commentId })
            );
        } else {
            res.json(useErrorResponseData('删除失败'));
        }
    } catch (e: unknown) {
        console.log('[api/comment/deleteComment] Error:', e);
        return res.json(useResponseData(50001, '删除失败，请稍后再试', null));
    }
};

const deleteComments = async () => {};

export { deleteComment, deleteComments };
