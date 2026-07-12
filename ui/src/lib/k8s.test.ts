import { describe, expect, it } from 'vitest'

import type { CustomResource } from '@/types/api'

import { getPrinterColumnValue } from './k8s'

const resource = {
  apiVersion: 'example.io/v1',
  kind: 'Widget',
  metadata: {
    name: 'example-widget',
    namespace: 'default',
  },
  spec: {
    http: [
      { match: { hosts: ['a.example', 'b.example'] } },
      { match: { hosts: ['c.example'] } },
    ],
  },
  status: {
    phase: 'Running',
    conditions: [
      { type: 'Synced', status: 'True' },
      { type: 'Ready', status: 'False' },
      { type: '[]', status: 'Empty Brackets' },
      { type: '[*]', status: 'Wildcard' },
    ],
    addresses: [
      { value: '10.0.0.1' },
      { value: null },
      {},
      { value: '10.0.0.2' },
    ],
  },
} as CustomResource

describe('getPrinterColumnValue', () => {
  it.each(['.status.phase', 'status.phase', '$.status.phase'])(
    'reads simple additionalPrinterColumns JSONPath values for %s',
    (jsonPath) => {
      expect(getPrinterColumnValue(resource, jsonPath)).toBe('Running')
    }
  )

  it('reads filtered conditions from additionalPrinterColumns JSONPath', () => {
    expect(
      getPrinterColumnValue(
        resource,
        ".status.conditions[?(@.type=='Synced')].status"
      )
    ).toBe('True')

    expect(
      getPrinterColumnValue(
        resource,
        ".status.conditions[?(@.type=='Ready')].status"
      )
    ).toBe('False')
  })

  it('does not normalize empty brackets inside filter strings', () => {
    expect(
      getPrinterColumnValue(
        resource,
        ".status.conditions[?(@.type=='[]')].status"
      )
    ).toBe('Empty Brackets')
  })

  it('returns undefined when the JSONPath does not match any value', () => {
    expect(
      getPrinterColumnValue(
        resource,
        ".status.conditions[?(@.type=='Healthy')].status"
      )
    ).toBeUndefined()
  })

  it('joins multiple values and skips nullish matches', () => {
    expect(getPrinterColumnValue(resource, '.status.addresses[*].value')).toBe(
      '10.0.0.1, 10.0.0.2'
    )
  })

  it('treats Kubernetes-style empty brackets as the first array item', () => {
    expect(getPrinterColumnValue(resource, '.status.addresses[].value')).toBe(
      '10.0.0.1'
    )
  })

  it('formats nested array values as JSON', () => {
    expect(getPrinterColumnValue(resource, '.spec.http[].match.hosts')).toBe(
      '["a.example","b.example"]'
    )
  })

  it('returns undefined instead of throwing for unparsable JSONPath', () => {
    expect(
      getPrinterColumnValue(resource, '.status.conditions[')
    ).toBeUndefined()
  })
})
