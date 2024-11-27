import express from 'express';
import {
    createComment,
    deleteComment,
    getComments,
    updateComment
} from '@nao-todo-server/apis';
// import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';

const exRouter: express.Router = express.Router();

exRouter.post('/comment', async (req, res) => {
    console.log('[/api/comment] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await createComment(req, res);
    // });
});

exRouter.delete('/comment', async (req, res) => {
    console.log('[/api/comment] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await deleteComment(req, res);
    // });
});

exRouter.put('/comment', async (req, res) => {
    console.log('[/api/comment] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await updateComment(req, res);
    // });
});

// exRouter.get('/comment', async (req, res) => {
//     console.log('[/api/comment] Route matched!', `(${req.method})`);
//     await connectAndRun(async () => {
//         await getComment(req, res);
//     });
// });

exRouter.get('/comments', async (req, res) => {
    console.log('[/api/comments] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getComments(req, res);
    // });
});

export default exRouter;
