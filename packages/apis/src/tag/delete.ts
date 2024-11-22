import { Tag } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';
import {
    getJWTPayload,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const deleteTag = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.tagId) {
            throw new Error('参数错误，请求无效');
        }

        const tagId = req.query.tagId as string;

        const deletedTag = await Tag.findOneAndDelete({
            _id: new ObjectId(tagId),
            userId
        });

        if (!deletedTag) throw new Error('删除失败');

        return res.json(
            useSuccessfulResponseData({ tagId: deletedTag._id.toString() })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/tag/deleteTag] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const deleteTags = async (req: Request, res: Response) => {};

export { deleteTag, deleteTags };
