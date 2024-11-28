import express from 'express';
import {
    createTag,
    deleteTag,
    getTag,
    getTags,
    updateTag
} from '@nao-todo-server/apis';
// import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';

const exRouter: express.Router = express.Router();

exRouter.post('/tag', async (req, res) => {
    console.log('[/api/tag] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await createTag(req, res);
    // });
});

exRouter.delete('/tag', async (req, res) => {
    console.log('[/api/tag] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await deleteTag(req, res);
    // });
});

exRouter.put('/tag', async (req, res) => {
    console.log('[/api/tag] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await updateTag(req, res);
    // });
});

exRouter.get('/tag', async (req, res) => {
    console.log('[/api/tag] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getTag(req, res);
    // });
});

exRouter.get('/tags', async (req, res) => {
    console.log('[/api/tags] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getTags(req, res);
    // });
});

export default exRouter;
