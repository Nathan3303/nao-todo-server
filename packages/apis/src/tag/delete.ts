import { Tag, Todo } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';
import {
    getJWTPayload,
    useErrorResponseData,
    useResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const deleteTag = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.query.tagId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
        }

        // 删除标签
        const deleteResult = await Tag.deleteOne({
            _id: new ObjectId(req.query.tagId as string),
            userId: new ObjectId(userId)
        });

        // 判断删除结果
        if (!deleteResult.deletedCount) {
            res.json(useErrorResponseData('删除失败'));
        }

        // 对所有待办移除标签记录
        const updateResult = await Todo.updateMany(
            {
                tags: { $in: [new ObjectId(req.query.tagId as string)] }
            },
            {
                $pull: { tags: new ObjectId(req.query.tagId as string) }
            }
        );

        // 响应
        res.json(
            useSuccessfulResponseData({
                tagId: req.query.tagId,
                effectedTodosCount: updateResult.modifiedCount
            })
        );
    } catch (e: unknown) {
        console.log('[api/tag/deleteTag] Error:', e);
        res.json(useResponseData(50001, '删除失败，请稍后再试', null));
    }
};

const deleteTags = async () => {};

export { deleteTag, deleteTags };
