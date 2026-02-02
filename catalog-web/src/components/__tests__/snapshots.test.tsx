import React from 'react'
import { render } from '@testing-library/react'

// Mock framer-motion to avoid issues in test environment
vi.mock('framer-motion', async () => ({
  motion: {
    div: React.forwardRef(({ children, ...props }: any, ref: any) => (
      <div ref={ref} {...props}>{children}</div>
    )),
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock lucide-react icons used across components
vi.mock('lucide-react', async () => {
  const icon = ({ className }: { className?: string }) => (
    <svg className={className} data-testid="mock-icon" />
  )
  return {
    ChevronDown: icon,
    Film: icon,
    Music: icon,
    Gamepad2: icon,
    Monitor: icon,
    BookOpen: icon,
    Star: icon,
    Calendar: icon,
    HardDrive: icon,
    Clock: icon,
    ExternalLink: icon,
    Download: icon,
    Eye: icon,
    Play: icon,
    Loader2: icon,
  }
})

import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import { Progress } from '@/components/ui/Progress'
import { Switch } from '@/components/ui/Switch'

describe('Snapshot Tests', () => {
  describe('Button', () => {
    it('renders default variant', () => {
      const { container } = render(<Button>Default</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders destructive variant', () => {
      const { container } = render(<Button variant="destructive">Delete</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders outline variant', () => {
      const { container } = render(<Button variant="outline">Outline</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders secondary variant', () => {
      const { container } = render(<Button variant="secondary">Secondary</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders ghost variant', () => {
      const { container } = render(<Button variant="ghost">Ghost</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders link variant', () => {
      const { container } = render(<Button variant="link">Link</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders disabled state', () => {
      const { container } = render(<Button disabled>Disabled</Button>)
      expect(container).toMatchSnapshot()
    })

    it('renders loading state', () => {
      const { container } = render(<Button loading>Loading</Button>)
      expect(container).toMatchSnapshot()
    })
  })

  describe('Badge', () => {
    it('renders default variant', () => {
      const { container } = render(<Badge>New</Badge>)
      expect(container).toMatchSnapshot()
    })

    it('renders secondary variant', () => {
      const { container } = render(<Badge variant="secondary">Draft</Badge>)
      expect(container).toMatchSnapshot()
    })

    it('renders destructive variant', () => {
      const { container } = render(<Badge variant="destructive">Error</Badge>)
      expect(container).toMatchSnapshot()
    })

    it('renders outline variant', () => {
      const { container } = render(<Badge variant="outline">Info</Badge>)
      expect(container).toMatchSnapshot()
    })
  })

  describe('Card', () => {
    it('renders full card with header, title, description, and content', () => {
      const { container } = render(
        <Card>
          <CardHeader>
            <CardTitle>My Card Title</CardTitle>
            <CardDescription>This is a description of the card.</CardDescription>
          </CardHeader>
          <CardContent>
            <p>Card body content goes here.</p>
          </CardContent>
        </Card>
      )
      expect(container).toMatchSnapshot()
    })
  })

  describe('Input', () => {
    it('renders normal input', () => {
      const { container } = render(
        <Input label="Username" placeholder="Enter username" />
      )
      expect(container).toMatchSnapshot()
    })

    it('renders input with error', () => {
      const { container } = render(
        <Input label="Email" error="Email is required" placeholder="Enter email" />
      )
      expect(container).toMatchSnapshot()
    })

    it('renders disabled input', () => {
      const { container } = render(
        <Input label="Locked field" disabled placeholder="Cannot edit" />
      )
      expect(container).toMatchSnapshot()
    })
  })

  describe('Select', () => {
    it('renders with options prop', () => {
      const { container } = render(
        <Select
          aria-label="Media type"
          options={[
            { value: 'movie', label: 'Movie' },
            { value: 'music', label: 'Music' },
            { value: 'game', label: 'Game' },
          ]}
          value="movie"
          onChange={() => {}}
        />
      )
      expect(container).toMatchSnapshot()
    })
  })

  describe('Progress', () => {
    it('renders at 0%', () => {
      const { container } = render(<Progress value={0} />)
      expect(container).toMatchSnapshot()
    })

    it('renders at 50%', () => {
      const { container } = render(<Progress value={50} />)
      expect(container).toMatchSnapshot()
    })

    it('renders at 100%', () => {
      const { container } = render(<Progress value={100} />)
      expect(container).toMatchSnapshot()
    })

    it('renders with label shown', () => {
      const { container } = render(<Progress value={75} showLabel />)
      expect(container).toMatchSnapshot()
    })
  })

  describe('Switch', () => {
    it('renders unchecked', () => {
      const { container } = render(
        <Switch checked={false} onCheckedChange={() => {}} />
      )
      expect(container).toMatchSnapshot()
    })

    it('renders checked', () => {
      const { container } = render(
        <Switch checked={true} onCheckedChange={() => {}} />
      )
      expect(container).toMatchSnapshot()
    })
  })
})
