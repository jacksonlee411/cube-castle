import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Outlet } from 'react-router-dom';
import { SidePanel } from '@workday/canvas-kit-react/side-panel';
import { space } from '@workday/canvas-kit-react/tokens';
import { Sidebar } from './Sidebar';
import { Header } from './Header';

const SIDE_PANEL_WIDTH = 312;

export const AppShell: React.FC = () => (
  <Box
    as="div"
    height="100vh"
    width="100vw"
    cs={{ display: 'flex', flexDirection: 'column' }}
  >
    <Header />
    <Box as="div" cs={{ display: 'flex', flex: 1, minHeight: 0 }}>
      <Box
        as="div"
        cs={{ position: 'relative', width: `${SIDE_PANEL_WIDTH}px`, flexShrink: 0, height: '100%' }}
      >
        <SidePanel
          open
          openWidth={SIDE_PANEL_WIDTH}
          backgroundColor={SidePanel.BackgroundColor.Gray}
          padding={space.m}
          header="导航"
          style={{ position: 'relative', width: '100%', height: '100%' }}
        >
          <Sidebar />
        </SidePanel>
      </Box>
      <Box as="main" cs={{ flex: 1, overflow: 'auto' }}>
        <Box as="div" padding="l">
          <Outlet />
        </Box>
      </Box>
    </Box>
  </Box>
);
