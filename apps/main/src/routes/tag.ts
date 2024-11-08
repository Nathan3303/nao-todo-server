import express from 'express';
import {
    createTag,
    deleteTag,
    updateTag,
    getTag,
    getTags
} from '@nao-todo-server/apis';

const exRouter: express.Router = express.Router();

exRouter.post('/tag', async (req, res) => {
    await createTag(req, res);
});

exRouter.delete('/tag', async (req, res) => {
    await deleteTag(req, res);
});

exRouter.put('/tag', async (req, res) => {
    await updateTag(req, res);
});

exRouter.get('/tag', async (req, res) => {
    await getTag(req, res);
});

exRouter.get('/tags', async (req, res) => {
    await getTags(req, res);
});

export default exRouter;
