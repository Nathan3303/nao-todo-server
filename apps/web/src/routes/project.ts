import express from 'express';
import {
    createProject,
    deleteProject,
    getProject,
    getProjects,
    updateProject
} from '@nao-todo-server/apis';
// import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';

const exRouter: express.Router = express.Router();

exRouter.post('/project', async (req, res) => {
    console.log('[/api/project] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await createProject(req, res);
    // });
});

exRouter.delete('/project', async (req, res) => {
    console.log('[/api/project] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await deleteProject(req, res);
    // });
});

exRouter.put('/project', async (req, res) => {
    console.log('[/api/project] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await updateProject(req, res);
    // });
});

exRouter.get('/project', async (req, res) => {
    console.log('[/api/project] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getProject(req, res);
    // });
});

exRouter.get('/projects', async (req, res) => {
    console.log('[/api/projects] Route matched!', `(${req.method})`);
    // await connectAndRun(async () => {
    await getProjects(req, res);
    // });
});

export default exRouter;
